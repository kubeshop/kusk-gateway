/*
The MIT License (MIT)

Copyright © 2022 Kubeshop

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
.
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Deprecated: `kusk api generate` is deprecated, use `kusk generate` instead.

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "parent command for api related functions",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Currently api only has one sub command
		fmt.Fprintln(os.Stderr, "The `api` command cannot be run directly. Please run: `kusk generate`")

		return cmd.Help()
	},
	// > For good practice, let's keep `kusk api generate` for a release and mark it as deprecated in the help. I would normally say show a deprecated message on usage but that would defeat the purpose as it would be a breaking change since we pipe the output.
	//
	// See: <https://github.com/kubeshop/kusk-gateway/issues/667>.
	Deprecated: "this command will be deprecated soon, please use `kusk generate`",
}

func init() {
	rootCmd.AddCommand(apiCmd)
}
