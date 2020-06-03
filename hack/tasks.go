package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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

const cdnURL = "https://stucco-release.fra1.digitaloceanspaces.com/%s/%s/%s/stucco%s"

func newVersion(bumpType versionBumpType) (v, nv semver.Version, err error) {
	_, err = gbtb.Output("git", "fetch", "--tags")
	if err != nil {
		return
	}
	o, err := gbtb.Output("git", "tag", "--sort=-taggerdate")
	if err != nil {
		return
	}
	var tag string
	reTag, err := regexp.Compile("^v[0-9]+\\.[0-9]+\\.[0-9]+$")
	if err != nil {
		return
	}
	for _, tg := range strings.Split(string(o), "\n") {
		tg = strings.TrimSpace(tg)
		if reTag.Match([]byte(tg)) {
			tag = tg
			break
		}
	}
	v, err = semver.Parse(tag[1:])
	if err != nil {
		return
	}
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
	if err != nil {
		return false, err
	}
	return len(o) == 0, nil
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
	darwinAmd64URL := fmt.Sprintf(cdnURL, to.String(), "darwin", "amd64", "")
	fmt.Fprintf(&b, "| macOS | [%s](%s) |\n", darwinAmd64URL, darwinAmd64URL)
	windowsAmd64URL := fmt.Sprintf(cdnURL, to.String(), "windows", "amd64", "")
	fmt.Fprintf(&b, "| windows | [%s](%s) |\n", windowsAmd64URL, windowsAmd64URL)
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
		[]string{"git", "add", "CHANGELOG.md"},
		[]string{"git", "commit", "-m", versionString},
		[]string{"git", "tag", versionString},
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

func generateProto() error {
	if err := os.Chdir("pkg/proto"); err != nil {
		return err
	}
	return gbtb.CommandJob("protoc", "-I", ".", "driver.proto", "--go_out=plugins=grpc:.")()
}

func coverage() error {

	args := []string{"-coverprofile=coverage.out", "./"}
	filepath.Walk("./pkg", func(path string, fi os.FileInfo, err error) error {
		if err != nil ||
			path == "./pkg" ||
			!fi.IsDir() ||
			filepath.Base(path) == "proto" ||
			strings.HasSuffix(filepath.Base(path), "test") {
			return err
		}
		args = append(args, "./"+path)
		return nil
	})
	return gbtb.Go("test", args...)()
}

func main() {
	gbtb.MustRun(
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
			Name: "generate-proto",
			Job:  generateProto,
		},
		&gbtb.Task{
			Name: "coverage",
			Job:  coverage,
		},
	)
}
