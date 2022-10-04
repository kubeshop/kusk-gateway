/*
The MIT License (MIT)

# Copyright © 2022 Kubeshop

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
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk-gateway/pkg/build"
)

func init() {
	versionCmd := NewVersionCommand(os.Stdout, build.Version)

	versionTemplate := "{{printf \"%s\" .Version}}\n"
	rootCmd.SetVersionTemplate(versionTemplate)

	formattedVersion := VersionFormat(build.Version)
	rootCmd.Version = formattedVersion

	rootCmd.AddCommand(versionCmd)
}

func NewVersionCommand(writer io.Writer, version string) *cobra.Command {
	formattedVersion := VersionFormat(version)

	return &cobra.Command{
		Use:   "version",
		Short: "version for Kusk",
		Run: func(*cobra.Command, []string) {
			fmt.Fprintf(writer, "%s\n", formattedVersion)
		},
	}
}

func VersionFormat(version string) string {
	version = strings.TrimPrefix(version, "v")

	return fmt.Sprintf("Kusk version %s\n%s", version, changelogURL(version))
}

func changelogURL(version string) string {
	path := "https://github.com/kubeshop/kusk-gateway"
	r := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[\w.]+)?$`)
	if !r.MatchString(version) {
		return fmt.Sprintf("%s/releases/latest", path)
	}

	url := fmt.Sprintf("%s/releases/tag/v%s", path, strings.TrimPrefix(version, "v"))
	return url
}
