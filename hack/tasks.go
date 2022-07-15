package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/Dennor/gbtb"
	"github.com/blang/semver/v4"
)

type versionBumpType uint

const (
	patch versionBumpType = iota
	minor
	major
)

const (
	cdnURL = "https://stucco-release.fra1.digitaloceanspaces.com/v%s/%s/%s/stucco%s"
)

var (
	reTag    = regexp.MustCompile("^v[0-9]+\\.[0-9]+\\.[0-9]+$")
	version  string
	dontTest = []string{
		"github.com/graphql-editor/stucco/hack",
		"github.com/graphql-editor/stucco/pkg/proto",
		"github.com/graphql-editor/stucco/pkg/proto/prototest",
	}
)

func semverParse(bv string) (semver.Version, error) {
	return semver.Parse(strings.TrimPrefix(bv, "v"))
}

func latestVersion() (v semver.Version, err error) {
	_, err = gbtb.Output("git", "fetch", "--tags")
	if err != nil {
		return
	}
	o, err := gbtb.Output("git", "tag", "--sort=-committerdate", "--merged")
	if err != nil {
		return
	}
	reTag, err := regexp.Compile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
	if err != nil {
		return
	}
	for _, tag := range strings.Split(string(o), "\n") {
		tag = strings.TrimSpace(tag)
		if reTag.Match([]byte(tag)) {
			v, err = semverParse(tag)
			break
		}
	}
	return
}

func newVersion(bumpType versionBumpType) (v, nv semver.Version, err error) {
	v, err = latestVersion()
	nv = v
	switch bumpType {
	case patch:
		if err = nv.IncrementPatch(); err != nil {
			return
		}
	case minor:
		if err = nv.IncrementMinor(); err != nil {
			return
		}
	case major:
		if err = nv.IncrementMajor(); err != nil {
			return
		}
	}
	return
}

func isClean() (bool, error) {
	o, err := gbtb.Output("git", "status", "--porcelain")
	return err == nil && len(o) == 0, err
}

func writeChangelog(from, to semver.Version) error {
	clean, err := isClean()
	if err != nil || !clean {
		if err == nil {
			err = fmt.Errorf("working directory is not clean")
		}
		return err
	}
	changelog, err := ioutil.ReadFile("CHANGELOG.md")
	if err != nil {
		return err
	}
	o, err := gbtb.Output(
		"git",
		"log",
		"--format=%h by %an: %s",
		fmt.Sprintf("v%s...HEAD", from.String()),
	)
	if err != nil {
		return err
	}
	lines := strings.Split(string(o), "\n")
	var b bytes.Buffer
	fmt.Fprintf(&b, "# Version v%s\n", to.String())
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "## Download")
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "|   | amd64 |")
	fmt.Fprintln(&b, "|---|----|")
	linuxAmd64URL := fmt.Sprintf(cdnURL, to.String(), "linux", "amd64", "")
	fmt.Fprintf(&b, "| linux | [%s](%s) |\n", linuxAmd64URL, linuxAmd64URL)
	linux386URL := fmt.Sprintf(cdnURL, to.String(), "linux", "386", "")
	fmt.Fprintf(&b, "| linux | [%s](%s) |\n", linux386URL, linux386URL)
	linuxArm64URL := fmt.Sprintf(cdnURL, to.String(), "linux", "arm64", "")
	fmt.Fprintf(&b, "| linux | [%s](%s) |\n", linuxArm64URL, linuxArm64URL)
	darwinAmd64URL := fmt.Sprintf(cdnURL, to.String(), "darwin", "amd64", "")
	fmt.Fprintf(&b, "| macOS | [%s](%s) |\n", darwinAmd64URL, darwinAmd64URL)
	darwinArm64URL := fmt.Sprintf(cdnURL, to.String(), "darwin", "arm64", "")
	fmt.Fprintf(&b, "| macOS | [%s](%s) |\n", darwinArm64URL, darwinArm64URL)
	windowsAmd64URL := fmt.Sprintf(cdnURL, to.String(), "windows", "amd64", "")
	fmt.Fprintf(&b, "| windows | [%s](%s) |\n", windowsAmd64URL, windowsAmd64URL)
	windows386URL := fmt.Sprintf(cdnURL, to.String(), "windows", "386", "")
	fmt.Fprintf(&b, "| windows | [%s](%s) |\n", windows386URL, windows386URL)
	windowsArm64URL := fmt.Sprintf(cdnURL, to.String(), "windows", "arm64", "")
	fmt.Fprintf(&b, "| windows | [%s](%s) |\n", windowsArm64URL, windowsArm64URL)
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "## Changes")
	fmt.Fprintln(&b, "")
	fmt.Fprintln(&b, "```")
	reFeatOrFix, err := regexp.Compile("^[^:]*: (feat|fix): ")
	if err != nil {
		return err
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if reFeatOrFix.Match([]byte(line)) {
			fmt.Fprintln(&b, line)
		}
	}
	fmt.Fprintln(&b, "```")
	fmt.Fprintln(&b, "")
	b.Write(changelog)
	return ioutil.WriteFile("CHANGELOG.md", b.Bytes(), os.ModePerm)
}

func commitAndTag(newVersion semver.Version) error {
	versionString := fmt.Sprintf("v%s", newVersion.String())
	for _, c := range [][]string{
		{"git", "add", "CHANGELOG.md"},
		{"git", "commit", "-m", versionString},
		{"git", "tag", versionString},
	} {
		_, err := gbtb.Output(c[0], c[1:]...)
		if err != nil {
			return err
		}
	}
	return nil
}

func versionBump(bumpType versionBumpType) func() error {
	return func() error {
		oldVersion, newVersion, err := newVersion(bumpType)
		if err != nil {
			return err
		}
		if err := writeChangelog(oldVersion, newVersion); err != nil {
			return err
		}
		return commitAndTag(newVersion)
	}
}

func testPackages() ([]string, error) {
	b, err := gbtb.Output("go", "list", "./...")
	if err != nil {
		return nil, err
	}
	var pkgs []string
	for _, pkg := range strings.Split(string(b), "\n") {
		if pkg != "" && !excludePkg(pkg) {
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs, nil
}

func runTests(coverage, race bool) error {
	pkgs, err := testPackages()
	if err != nil {
		return err
	}
	args := []string{"test"}
	if coverage {
		args = append(args, "-coverprofile=coverage.out")
	}
	if race {
		args = append(args, "-race")
	}
	args = append(args, pkgs...)
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return gbtb.PipeCommands(cmd)
}

func test() error {
	return runTests(false, false)
}

func testRace() error {
	return runTests(false, true)
}

func coverage() (err error) {
	return runTests(true, false)
}

func out(s string) string {
	return filepath.Join("bin", s)
}

var (
	goBuildDependencies = gbtb.DependenciesList{
		gbtb.GlobFiles("**/*.go"),
		gbtb.StaticDependencies{"go.sum", "go.mod"},
	}
)

type flavour struct {
	goos, goarch, out, ext string
	azureFunction          bool
}

func buildVersion() string {
	if version != "" {
		return version
	}
	// if directory is not clean, leave build version empty
	b, err := isClean()
	if err == nil && b {
		// check current HEAD ref and use it as a version
		// unless it's tagged with a version tag
		o, err := gbtb.Output("git", "show-ref", "--head", "HEAD")
		if err == nil {
			hRef := strings.Split(string(o), " ")[0]
			buildVersion := hRef[:12]
			v, err := latestVersion()
			if err == nil {
				ver := "v" + v.String()
				o, err := gbtb.Output("git", "show-ref", ver)
				if err == nil {
					// check if latest version tag is equal to current HEAD
					if hRef == strings.Split(string(o), " ")[0] {
						buildVersion = ver
					}
				}
			}
			return buildVersion
		}
	}
	return ""
}

func ldflags(bv string) string {
	return fmt.Sprintf("-ldflags=-X github.com/graphql-editor/stucco/pkg/version.BuildVersion=%s", bv)
}

func xBuildCommandLine(f flavour, bv string) func() error {
	return func() error {
		opts := []string{"build", "-o", f.out}
		if bv != "" {
			opts = append(opts, ldflags(bv))
		}
		opts = append(opts, "./main.go")
		cmd := exec.Command("go", opts...)
		cmd.Env = append(cmd.Env, append([]string{
			"GOARCH=" + f.goarch,
			"GOOS=" + f.goos,
		})...)
		if f.goos != "darwin" {
			cmd.Env = append(cmd.Env, "CGO_ENABLED=0")
		} else {
			cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
		}
		return gbtb.PipeCommands(cmd)
	}
}

type zipData struct {
	dst, src string
}

func zipFiles(filename string, files []zipData) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err = addFileToZip(zipWriter, file.dst, file.src); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, dst, src string) error {
	fileToZip, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = dst
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func makeAzureFunctionZip(flavours []flavour) func() error {
	return func() error {
		azureOut := filepath.Join("bin", "azure")
		err := os.MkdirAll(azureOut, 0755)
		if err != nil {
			return err
		}
		files := []zipData{}
		for _, funcFile := range []string{
			filepath.Join("graphql", "function.json"),
			filepath.Join("webhook", "function.json"),
			"host.json",
			"run.js",
		} {
			files = append(
				files,
				zipData{
					dst: funcFile,
					src: filepath.Join("pkg", "providers", "azure", "function", funcFile),
				},
			)
		}
		for _, fv := range flavours {
			if fv.azureFunction {
				files = append(
					files,
					zipData{
						dst: filepath.Join("stucco", fv.goos, fv.goarch, "stucco"+fv.ext),
						src: fv.out,
					},
				)
			}
		}
		return zipFiles(filepath.Join(azureOut, "function.zip"), files)
	}
}

func helpTask(tasks *gbtb.Tasks) gbtb.Job {
	return func() error {
		for _, t := range *tasks {
			names := t.GetNames()
			fmt.Printf("%s: %s\n", names[0], strings.Join(names, ","))
		}
		return nil
	}
}

func conditionalJob(cond func() bool, j gbtb.Job) gbtb.Job {
	return func() (err error) {
		if cond() {
			err = j()
		}
		return
	}
}

func majorString(bv string) string {
	v, err := semverParse(bv)
	if err != nil {
		return "v9999"
	}
	return fmt.Sprintf("v%d", v.Major)
}

func minorString(bv string) string {
	v, err := semverParse(bv)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("v%d.%d", v.Major, v.Minor)
}

func isVersionCond(bv string) func() bool {
	return func() bool {
		_, err := semverParse(bv)
		return err == nil
	}
}

func excludePkg(pkg string) bool {
	for _, exclude := range dontTest {
		if pkg == exclude {
			return true
		}
	}
	return false
}

func main() {
	gbtb.FlagsInit(flag.CommandLine)
	flag.StringVar(&version, "version", "", "build version")
	flag.Parse()
	bv := buildVersion()
	tasks := gbtb.Tasks{
		&gbtb.Task{
			Name: "bump-patch",
			Job:  versionBump(patch),
		},
		&gbtb.Task{
			Name: "bump-minor",
			Job:  versionBump(minor),
		},
		&gbtb.Task{
			Name: "bump-major",
			Job:  versionBump(major),
		},
		&gbtb.Task{
			Name: "test",
			Job:  test,
		},
		&gbtb.Task{
			Name: "test-race",
			Job:  testRace,
		},
		&gbtb.Task{
			Name: "coverage",
			Job:  coverage,
		},
	}
	functionFlavours := []flavour{
		{goos: "linux", goarch: "amd64", azureFunction: true},
		{goos: "linux", goarch: "386", azureFunction: true},
		{goos: "linux", goarch: "arm64", azureFunction: true},
		{goos: "windows", goarch: "amd64", ext: ".exe", azureFunction: true},
		{goos: "windows", goarch: "386", ext: ".exe", azureFunction: true},
		{goos: "windows", goarch: "arm64", ext: ".exe", azureFunction: true},
	}
	cliFlavours := append([]flavour{}, functionFlavours...)
	if runtime.GOOS == "darwin" {
		cliFlavours = append(cliFlavours, []flavour{
			{goos: "darwin", goarch: "amd64"},
			{goos: "darwin", goarch: "arm64"},
		}...)
	}
	for i, f := range functionFlavours {
		functionFlavours[i].out = out(filepath.Join("cli", f.goos, f.goarch, "stucco"+f.ext))
	}
	for i, f := range cliFlavours {
		cliFlavours[i].out = out(filepath.Join("cli", f.goos, f.goarch, "stucco"+f.ext))
	}
	var cliDeps gbtb.StaticDependencies
	for _, f := range cliFlavours {
		cliDeps = append(cliDeps, f.out)
		// keep job names consitent across operating systems
		name := strings.Join([]string{"bin", "cli", f.goos, f.goarch, "stucco" + f.ext}, "/")
		tasks = append(tasks, &gbtb.Task{
			Name:         name,
			Job:          xBuildCommandLine(f, bv),
			Dependencies: goBuildDependencies,
		})
	}
	tasks = append(
		tasks, &gbtb.Task{
			Name:         "build_cli",
			Dependencies: cliDeps,
		},
		&gbtb.Task{
			Name:         "build_azure_function",
			Dependencies: cliDeps,
			Job:          makeAzureFunctionZip(functionFlavours),
		},
		&gbtb.Task{
			Name: "help",
			Job:  helpTask(&tasks),
		},
	)
	tasks.Do(flag.Args()...)
}
