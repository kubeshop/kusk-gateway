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
*/
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/kuskui"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
	"github.com/kubeshop/kusk-gateway/pkg/analytics"
	"github.com/kubeshop/kusk-gateway/pkg/build"
	"github.com/mattn/go-isatty"
)

var cfgFile string
var verbose bool

const (
	kuskgateway          = "kusk-gateway"
	kusknamespace        = "kusk-system"
	kuskgatewayapi       = "kusk-gateway-api"
	kuskgatewaydashboard = "kusk-gateway-dashboard"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kusk",
	Short: "",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		analytics.SendAnonymousCMDInfo(nil)

		if isatty.IsTerminal(os.Stdout.Fd()) == true &&
			build.Version != "latest" {

			if len(build.Version) != 0 {
				ghclient, err := utils.NewGithubClient("", nil)
				if err != nil {
					errors.NewErrorReporter(cmd, err).Report()
					return
				}

				ref, err := ghclient.GetLatest(kuskgateway)
				if err != nil {
					errors.NewErrorReporter(cmd, err).Report()
					return
				}

				latestVersion, err := version.NewVersion(ref)
				if err != nil {
					errors.NewErrorReporter(cmd, err).Report()
					return
				}

				currentVersion, err := version.NewVersion(build.Version)
				if err != nil {
					errors.NewErrorReporter(cmd, err).Report()
					return
				}

				if currentVersion != nil && currentVersion.LessThan(latestVersion) {
					kuskui.PrintWarning(fmt.Sprintf("This version %s of Kusk cli is outdated. The latest version available is %s\n", currentVersion, latestVersion), "Please follow instructions to update you installation: https://docs.kusk.io/reference/cli/overview/#updating")
					return
				}
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		errors.NewErrorReporter(rootCmd, err).Report()
	}

	if err != nil {
		kuskui.PrintError(err.Error())
		os.Exit(1)
	}
}

const (
	cmdGroupAnnotation = "GroupAnnotation"
	cmdMngmCmdGroup    = "1-Management commands"
	cmdGroupCommands   = "2-Commands"
	cmdGroupCobra      = "other"

	cmdGroupDelimiter = "-"
)

func helpMessageByGroups(cmd *cobra.Command) string {

	groups := map[string][]string{}
	for _, c := range cmd.Commands() {
		var groupName string
		v, ok := c.Annotations[cmdGroupAnnotation]
		if !ok {
			groupName = cmdGroupCobra
		} else {
			groupName = v
		}

		groupCmds := groups[groupName]
		groupCmds = append(groupCmds, fmt.Sprintf("%-16s%s", c.Name(), kuskui.Gray(c.Short)))
		sort.Strings(groupCmds)

		groups[groupName] = groupCmds
	}

	if len(groups[cmdGroupCobra]) != 0 {
		groups[cmdMngmCmdGroup] = append(groups[cmdMngmCmdGroup], groups[cmdGroupCobra]...)
	}
	delete(groups, cmdGroupCobra)

	groupNames := []string{}
	for k, _ := range groups {
		groupNames = append(groupNames, k)
	}
	sort.Strings(groupNames)

	buf := bytes.Buffer{}
	for _, groupName := range groupNames {
		commands := groups[groupName]

		groupSplit := strings.Split(groupName, cmdGroupDelimiter)
		group := "others"
		if len(groupSplit) > 1 {
			group = strings.Split(groupName, cmdGroupDelimiter)[1]
		}
		buf.WriteString(fmt.Sprintf("%s\n", kuskui.Gray(group)))

		for _, cmd := range commands {
			buf.WriteString(fmt.Sprintf("%s\n", cmd))
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".kusk" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kusk")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("failed to read config file %q, %w", viper.ConfigFileUsed(), err))
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.SetHelpFunc(help)
}

func help(c *cobra.Command, s []string) {

	switch c.Use {
	case mockCmd.Use:
		fmt.Println("")
		mockDescription = strings.Replace(mockDescription, "Description:", kuskui.Gray("Description:"), 1)
		mockHelp = strings.Replace(mockHelp, "Schema example:", kuskui.Gray("Schema example:"), 1)
		mockHelp = strings.Replace(mockHelp, "Generated JSON Response:", kuskui.Gray("Generated JSON Response:"), 1)
		mockHelp = strings.Replace(mockHelp, "Generated XML Response:", kuskui.Gray("Generated XML Response:"), 1)
		mockHelp = strings.Replace(mockHelp, "XML Respose from Defined Examples:", kuskui.Gray("XML Respose from Defined Examples:"), 1)
		mockHelp = strings.Replace(mockHelp, "Stop Mock Server:", kuskui.Gray("Stop Mock Server:"), 1)

		fmt.Println(mockDescription)
		fmt.Println(mockHelp)
		fmt.Println("")
	case generateCmd.Use:
		fmt.Println("")
		generateDescription = strings.Replace(generateDescription, "Description:", kuskui.Gray("Description:"), 1)
		generateHelp = strings.Replace(generateHelp, "No name specified:", kuskui.Gray("No name specified::"), 1)
		generateHelp = strings.Replace(generateHelp, "No API Name Specified:", kuskui.Gray("No API Name Specified:"), 1)
		generateHelp = strings.Replace(generateHelp, "Namespace Specified:", kuskui.Gray("Namespace Specified:"), 1)
		generateHelp = strings.Replace(generateHelp, "OpenAPI Definition from URL:", kuskui.Gray("OpenAPI Definition from URL:"), 1)

		fmt.Println(generateDescription)
		fmt.Println(generateHelp)
		fmt.Println("")
	default:
		if len(c.Short) != 0 {
			fmt.Println("")
			kuskui.PrintInfo(c.Short)
			fmt.Println("")
		}
	}

	kuskui.PrintInfoGray("Usage")
	kuskui.PrintInfo(fmt.Sprintf("%s %s", c.Use, kuskui.Gray("[flags]")))
	if len(c.Commands()) > 0 {
		kuskui.PrintInfo(fmt.Sprintf("%s %s", c.Use, kuskui.Gray("[command]")))
	}

	fmt.Println("")
	usage := helpMessageByGroups(c)
	kuskui.PrintInfo(usage)
	kuskui.PrintInfoGray(kuskui.Gray("Flags"))
	kuskui.PrintInfo(c.Flags().FlagUsages())
	kuskui.PrintInfo("Use \"kusk [command] --help\" for more information about a command.")
	fmt.Println("")
	kuskui.PrintInfo(fmt.Sprintf("%s   %s", kuskui.Gray("Docs & Support:"), "https://docs.kusk.io/"))
	fmt.Println("")

}
