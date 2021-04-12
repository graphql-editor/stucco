package runtimes

import "io"

// OsType defines target os in Azure Functions
type OsType uint8

// List of operating systems supported by azure function
const (
	Linux OsType = iota
	Windows
)

var commonIgnoreList = []string{"/.*"}

// File included in function generation
type File struct {
	io.Reader
	Path string
}
