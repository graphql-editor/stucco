// +build windows

package plugin

import (
	"os"
	"strings"
)

var extecutableExtensions = []string{
	".exe",
	".cmd",
	".bat",
}

func isExecutable(fi os.FileInfo) bool {
	for _, ext := range extecutableExtensions {
		if strings.HasSuffix(fi.Name(), ext) {
			return true
		}
	}
	return false
}
