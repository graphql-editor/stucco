// +build !windows

package plugin

import "os"

func isExecutable(fi os.FileInfo) bool {
	return fi.Mode()&0111 != 0
}
