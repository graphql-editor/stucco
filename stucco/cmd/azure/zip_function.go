package azurecmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/graphql-editor/stucco/pkg/providers/azure/project/runtimes"
	"github.com/graphql-editor/stucco/pkg/router"
	"github.com/graphql-editor/stucco/pkg/server"
	"github.com/graphql-editor/stucco/pkg/types"
	"github.com/graphql-editor/stucco/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

type functionRuntime uint8

func (f *functionRuntime) Type() string {
	return "functionRuntime"
}

func (f *functionRuntime) Set(v string) error {
	switch v {
	case "node":
		*f = node
	default:
		return fmt.Errorf("%s is not a supported value for function runtime", v)
	}
	return nil
}

func (f *functionRuntime) String() string {
	switch *f {
	case node:
		return "node"
	}
	return "<invalid>"
}

const (
	node functionRuntime = iota
	invalid
)

// runtimeConfig represents runtime configuration used while generating function bundle
type runtimeConfig interface {
	Function(f types.Function) ([]runtimes.File, error)
	IgnoreFiles() []string
	GlobalFiles() ([]runtimes.File, error)
}

var commonSkip = []string{".git"}
var rtConfigs = []runtimeConfig{
	runtimes.StuccoJS{},
}

func appendRuntimeFiles(files []runtimes.File, w *zip.Writer, projectDir, prefix string) {
	for _, f := range files {
		p := f.Path
		if prefix != "" {
			p = filepath.Join(filepath.Dir(p), prefix+filepath.Base(p))
		}
		if _, err := os.Stat(filepath.Join(projectDir, p)); err == nil {
			klog.Warningf("skipping %s becuase it exists", p)
			continue
		}
		zd := utils.ZipData{
			Data:     f.Reader,
			Filename: p,
		}
		if err := zd.AddFileToZip(w); err != nil {
			klog.Fatal(err)
		}
	}
}

type functionList interface {
	functions() []types.Function
}

type interfaceResolveTypeFunctionList map[string]router.InterfaceConfig

func (i interfaceResolveTypeFunctionList) functions() []types.Function {
	ret := make([]types.Function, len(i))
	for _, v := range i {
		ret = append(ret, v.ResolveType)
	}
	return ret
}

type resolveFunctionList map[string]router.ResolverConfig

func (r resolveFunctionList) functions() []types.Function {
	ret := make([]types.Function, len(r))
	for _, v := range r {
		ret = append(ret, v.Resolve)
	}
	return ret
}

type scalarsFunctionList map[string]router.ScalarConfig

func (s scalarsFunctionList) functions() []types.Function {
	ret := make([]types.Function, len(s))
	for _, v := range s {
		ret = append(ret, v.Serialize)
		ret = append(ret, v.Parse)
	}
	return ret
}

type unionFunctionList map[string]router.UnionConfig

func (u unionFunctionList) functions() []types.Function {
	ret := make([]types.Function, len(u))
	for _, v := range u {
		ret = append(ret, v.ResolveType)
	}
	return ret
}

type subscriptionFunctionList map[string]router.SubscriptionConfig

func (s subscriptionFunctionList) functions() []types.Function {
	ret := make([]types.Function, len(s))
	for _, v := range s {
		ret = append(ret, v.CreateConnection)
		ret = append(ret, v.Listen)
	}
	return ret
}

func appendAllRuntimeFiles(cfg router.Config, rt runtimeConfig, w *zip.Writer, projectDir, prefix string) {
	functions := interfaceResolveTypeFunctionList(cfg.Interfaces).functions()
	functions = append(functions, resolveFunctionList(cfg.Resolvers).functions()...)
	functions = append(functions, scalarsFunctionList(cfg.Scalars).functions()...)
	functions = append(functions, unionFunctionList(cfg.Unions).functions()...)
	functions = append(functions, subscriptionFunctionList(map[string]router.SubscriptionConfig{
		"": cfg.Subscriptions,
	}).functions()...)
	functions = append(functions, subscriptionFunctionList(cfg.SubscriptionConfigs).functions()...)
	for i := 0; i < len(functions); i++ {
		f := functions[i]
		// Remove empty functions and repetitions
		if f.Name == "" || sort.Search(i, func(i int) bool {
			return f.Name == functions[i].Name
		}) != i {
			functions = append(functions[:i], functions[i+1:]...)
			i--
		}
	}
	globalFiles, err := rt.GlobalFiles()
	if err != nil {
		klog.Fatal(err)
	}
	appendRuntimeFiles(globalFiles, w, projectDir, "")
	for _, f := range functions {
		funcFiles, err := rt.Function(f)
		if err != nil {
			klog.Fatal(err)
		}
		appendRuntimeFiles(funcFiles, w, projectDir, prefix)
	}
}

// NewZipFunctionCommand returns new zip-function command
func NewZipFunctionCommand() *cobra.Command {
	var config string
	var ca string
	var output string
	var prefix string
	var insecure bool
	runtime := node
	var projectDir string
	zipFunction := &cobra.Command{
		Use:   "zip-function",
		Short: "Create function zip that can be used in azcli to deploy function",
		Run: func(cmd *cobra.Command, args []string) {
			var cfg server.Config
			err := utils.LoadConfigFile(config, &cfg)
			if err != nil {
				klog.Fatal(err)
			}
			var extraFiles []utils.ZipData
			if filepath.Clean(ca) != "ca.pem" {
				caData, err := utils.ReadLocalOrRemoteFile(ca)
				if err != nil && !insecure {
					klog.Fatal(err)
				}
				if caData != nil {
					extraFiles = append(extraFiles, utils.ZipData{Filename: "ca.pem", Data: bytes.NewReader(caData)})
				}
			} else if !insecure {
				fi, err := os.Stat("ca.pem")
				if err != nil || fi.IsDir() {
					klog.Fatal("enable insecure or provide certificate authority")
				}
			}
			if runtime >= invalid {
				klog.Fatal("invalid runtime")
			}
			if projectDir == "" {
				projectDir = "."
			}
			d := filepath.Dir(output)
			if d != "" {
				err = os.MkdirAll(d, 0755)
				if err != nil {
					klog.Fatal(err)
				}
			}
			newZipFile, err := os.Create(output)
			if err != nil {
				klog.Fatal(err)
			}
			rtConfig := rtConfigs[int(runtime)]
			defer newZipFile.Close()
			zipWriter := zip.NewWriter(newZipFile)
			defer zipWriter.Close()
			if err := utils.AddPathToZip(projectDir, rtConfig.IgnoreFiles(), zipWriter); err != nil {
				klog.Fatal(err)
			}
			for _, ef := range extraFiles {
				if err := ef.AddFileToZip(zipWriter); err != nil {
					klog.Fatal(err)
				}
			}
			appendAllRuntimeFiles(cfg.Config, rtConfig, zipWriter, projectDir, prefix)
		},
	}
	zipFunction.Flags().StringVarP(&config, "config", "c", os.Getenv("STUCCO_CONFIG"), "Path or url to stucco config")
	zipFunction.Flags().StringVar(&ca, "ca", "ca.pem", "ca used to sign router certificate")
	zipFunction.Flags().StringVarP(&output, "out", "o", "dist/function.zip", "Function archive output")
	zipFunction.Flags().BoolVarP(&insecure, "insecure", "i", false, "Allow zip without ca file")
	zipFunction.Flags().VarP(&runtime, "runtime", "r", "Function target runtime")
	zipFunction.Flags().StringVarP(&projectDir, "project-dir", "p", "", "Project directory, defaults to current directory")
	zipFunction.Flags().StringVar(&prefix, "prefix", "", "Optional prefix name to generated function directory names")
	return zipFunction
}
