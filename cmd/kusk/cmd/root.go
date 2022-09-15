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

*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kubeshop/testkube/pkg/ui"

	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/errors"
	"github.com/kubeshop/kusk-gateway/cmd/kusk/internal/utils"
	"github.com/kubeshop/kusk-gateway/pkg/analytics"
	"github.com/kubeshop/kusk-gateway/pkg/build"
)

var (
	cfgFile string
	// verbosity - controls verbosity level, 0 is the lowest and results in the lowest amount of information being printed.
	verbosityLevel uint32
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kusk",
	Short: "",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		analytics.SendAnonymousCMDInfo(nil)
		if cmd.Name() != generateCmd.Name() {

			if len(build.Version) != 0 {
				ghclient, _ := utils.NewGithubClient("", nil)
				i, _, err := ghclient.GetTags()
				if err != nil {
					errors.NewErrorReporter(cmd, err).Report()
				}

				if len(i) > 0 {
					ref_str := strings.Split(i[len(i)-1].Ref, "/")
					ref := ref_str[len(ref_str)-1]

					latestVersion, err := version.NewVersion(ref)
					if err != nil {
						errors.NewErrorReporter(cmd, err).Report()
					}

					currentVersion, err := version.NewVersion(build.Version)
					if err != nil {
						errors.NewErrorReporter(cmd, err).Report()
					}

					if currentVersion.LessThan(latestVersion) {
						ui.Warn(fmt.Sprintf("This version %s of Kusk cli is outdated. The latest version available is %s\n", currentVersion, latestVersion), "Please follow instructions to update you installation: https://docs.kusk.io/reference/cli/overview/#updating")
					}
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

	cobra.CheckErr(err)
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kusk.yaml)")
	rootCmd.PersistentFlags().Uint32VarP(&verbosityLevel, "verbosity", "v", 0, "number for the log level verbosity, 0=error, 1=warn, 2=info, 3=debug and 4=trace.")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

	// If running at debug level, turn on helm debugging.
	// Available environment variables can be found by running `helm env`.
	if isLevelDebug() {
		const HelmDebugKey = "HELM_DEBUG"
		const HelmDebugValue = "true"
		if err := os.Setenv(HelmDebugKey, HelmDebugValue); err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("initialisation error: unable to set environmental variable `%v=%q`: %w", HelmDebugKey, HelmDebugValue, err))
			os.Exit(1)
		}
	}
}

func isLevelError() bool {
	return verbosityLevel == 0
}

func isLevelWarn() bool {
	return verbosityLevel == 1
}

func isLevelInfo() bool {
	return verbosityLevel == 2
}

func isLevelDebug() bool {
	return verbosityLevel == 3
}

func isLevelTrace() bool {
	return verbosityLevel == 4
}

func getHelmCommandArguments(arguments ...string) []string {
	commandArguments := []string{}
	if isLevelDebug() {
		commandArguments = append([]string{"--debug"}, arguments...)
	} else {
		commandArguments = append(commandArguments, arguments...)
	}
	return commandArguments
}
