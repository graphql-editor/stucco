package azurecmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/graphql-editor/stucco/pkg/providers/azure/project"
	"github.com/graphql-editor/stucco/pkg/providers/azure/vars"
	"github.com/graphql-editor/stucco/pkg/utils"
	global_vars "github.com/graphql-editor/stucco/pkg/vars"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func webhookJSON(whPattern string) io.Reader {
	return strings.NewReader(`{
  "bindings": [
    {
      "authLevel": "Anonymous",
      "type": "httpTrigger",
      "direction": "in",
      "name": "req",
      "methods": ["get", "post"],
	 "route": "` + whPattern + `"
    },
    {
      "type": "http",
      "direction": "out",
      "name": "res"
    }
  ]
}`)
}

// NewZipRouterCommand returns new zip-router command
func NewZipRouterCommand() *cobra.Command {
	var config string
	var schema string
	var cert string
	var key string
	var output string
	var insecure bool
	var ver string
	var host string
	zipRouter := &cobra.Command{
		Use:   "zip-router",
		Short: "Create router function zip that can be used in azcli to deploy function",
		Run: func(cmd *cobra.Command, args []string) {
			configData, err := utils.ReadLocalOrRemoteFile(config)
			if err != nil {
				klog.Fatal(err)
			}
			schemaData, err := utils.ReadLocalOrRemoteFile(schema)
			if err != nil {
				klog.Fatal(err)
			}
			keyData, err := utils.ReadLocalOrRemoteFile(key)
			if err != nil && !insecure {
				klog.Fatal(err)
			}
			certData, err := utils.ReadLocalOrRemoteFile(cert)
			if err != nil && !insecure {
				klog.Fatal(err)
			}
			var cfg project.Config
			if err := utils.LoadConfigFile(config, &cfg); err != nil {
				klog.Fatal(err)
			}
			extraFiles := []utils.ZipData{
				{Filename: "stucco.json", Data: bytes.NewReader(configData)},
				{Filename: "schema.graphql", Data: bytes.NewReader(schemaData)},
			}
			for i, wh := range cfg.AzureOpts.Webhooks {
				extraFiles = append(extraFiles, utils.ZipData{
					Filename: "webhook" + strconv.FormatInt(int64(i), 10) + "/function.json",
					Data:     webhookJSON(wh),
				})
			}
			if keyData != nil {
				extraFiles = append(extraFiles, utils.ZipData{Filename: "key.pem", Data: bytes.NewReader(keyData)})
			}
			if certData != nil {
				extraFiles = append(extraFiles, utils.ZipData{Filename: "cert.pem", Data: bytes.NewReader(certData)})
			}
			var r project.Router
			if ver != "" || host != "" {
				r.Vars = &vars.Vars{
					Vars: global_vars.Vars{
						Relase: global_vars.Release{
							Version: ver,
							Host:    host,
						},
					},
				}
			}
			rc, err := r.Zip(extraFiles)
			if err != nil {
				klog.Fatal(err)
			}
			defer rc.Close()
			d := filepath.Dir(output)
			if d != "" {
				err = os.MkdirAll(d, 0755)
				if err != nil {
					klog.Fatal(err)
				}
			}
			f, err := os.Create(output)
			if err != nil {
				klog.Fatal(err)
			}
			defer f.Close()
			_, err = io.Copy(f, rc)
			if err != nil {
				klog.Fatal(err)
			}
		},
	}
	defaultConfig := os.Getenv("STUCCO_CONFIG")
	if defaultConfig == "" {
		defaultConfig = "./stucco.json"
	}
	defaultSchema := os.Getenv("STUCCO_SCHEMA")
	if defaultSchema == "" {
		defaultSchema = "./schema.graphql"
	}
	zipRouter.Flags().StringVarP(&config, "config", "c", defaultConfig, "Path or url to stucco config")
	zipRouter.Flags().StringVarP(&schema, "schema", "s", defaultSchema, "Path or url to stucco schema")
	zipRouter.Flags().StringVar(&key, "key", "key.pem", "key used in http client cert authentication")
	zipRouter.Flags().StringVar(&cert, "cert", "cert.pem", "cert used in http client cert authentication")
	zipRouter.Flags().StringVarP(&output, "out", "o", "dist/router.zip", "Router function archive output")
	zipRouter.Flags().StringVar(&ver, "zip-version", "", "Use specific version of zip as a base")
	zipRouter.Flags().StringVar(&host, "zip-host", "", "Override router base zip host")
	zipRouter.Flags().BoolVarP(&insecure, "insecure", "i", false, "Allow zip without certificate files")
	return zipRouter
}
