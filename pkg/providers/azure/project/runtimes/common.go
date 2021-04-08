package runtimes

import "io"

var commonIgnoreList = []string{"/.*"}

// File included in function generation
type File struct {
	io.Reader
	Path string
}
