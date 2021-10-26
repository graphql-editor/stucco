/*
Copyright © 2020 Kamil Matysiewicz <kamil@graphqleditor.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"fmt"
	"os"
	"time"

	crypto_rand "crypto/rand"
	"encoding/binary"
	math_rand "math/rand"

	"github.com/graphql-editor/stucco/cmd"
	azurecmd "github.com/graphql-editor/stucco/cmd/azure"
	localcmd "github.com/graphql-editor/stucco/cmd/local"
	"github.com/spf13/cobra"

	configcmd "github.com/graphql-editor/stucco/cmd/config"
)

func seed() {
	// seed basic rand source
	// try seed it with secure seed from crypto, if it fails, fallback
	// to time
	seed := time.Now().UTC().UnixNano()
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err == nil {
		seed = int64(binary.LittleEndian.Uint64(b[:]))
	}
	math_rand.Seed(seed)
}

func main() {
	seed()
	rootCmd := &cobra.Command{
		Use:   "stucco",
		Short: "Set of tools to work with stucco",
	}
	rootCmd.AddCommand(cmd.NewVersionCommand())
	rootCmd.AddCommand(azurecmd.NewAzureCommand())
	rootCmd.AddCommand(localcmd.NewLocalCommand())
	rootCmd.AddCommand(configcmd.NewConfigCommand())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
