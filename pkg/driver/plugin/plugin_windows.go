// +build windows

package plugin

import (
	"os"
	"strings"
)

func isExecutable(fi os.FileInfo) bool {
	return strings.HasSuffix(fi.Name(), ".exe")
}
