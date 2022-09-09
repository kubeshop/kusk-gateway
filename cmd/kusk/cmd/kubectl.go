/*
Copyright 2022 k0s authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	// "github.com/k0sproject/k0s/pkg/config"
	// "github.com/k0sproject/k0s/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/component-base/logs"
	kubectl "k8s.io/kubectl/pkg/cmd"
)

// type CmdOpts config.CLIOptions

type kubectlPluginHandler struct{}

func (h *kubectlPluginHandler) Lookup(filename string) (string, bool) {
	path, err := exec.LookPath(fmt.Sprintf("kubectl-%s", filename))
	if err != nil || path == "" {
		return "", false
	}
	return path, true
}

// adapted from kubectl.DefaultPluginHandler
func (h *kubectlPluginHandler) Execute(executablePath string, cmdArgs, environment []string) error {
	if _, err := exec.LookPath("kubectl"); err != nil {
		if exe, err := os.Executable(); err == nil {
			logrus.Warn("kubectl not found in $PATH. many kubectl plugins try to run 'kubectl'.`", exe)
		}
	}

	// Windows does not support exec syscall.
	if runtime.GOOS == "windows" {
		cmd := exec.Command(executablePath, cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = environment
		if err := cmd.Run(); err != nil {
			return err
		}
		os.Exit(0)
	}

	// invoke cmd binary relaying the environment and args given
	// append executablePath to cmdArgs, as execve will make first argument the "binary name".
	return syscall.Exec(executablePath, append([]string{executablePath}, cmdArgs...), environment)
}

// func init() {
// 	rootCmd.AddCommand(NewKubectlCmd())
// }

func NewKubectlCmd() *cobra.Command {
	_ = pflag.CommandLine.MarkHidden("log-flush-frequency")
	_ = pflag.CommandLine.MarkHidden("version")

	args := kubectl.KubectlOptions{
		IOStreams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		Arguments:     os.Args,
		PluginHandler: &kubectlPluginHandler{},
	}
	cmd := kubectl.NewKubectlCommand(args)

	cmd.Aliases = []string{"kc"}
	cmd.Hidden = true
	// Get handle on the original kubectl prerun so we can call it later
	originalPreRunE := cmd.PersistentPreRunE
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Call parents pre-run if exists, cobra does not do this automatically
		// See: https://github.com/spf13/cobra/issues/216
		if parent := cmd.Parent(); parent != nil {
			if parent.PersistentPreRun != nil {
				parent.PersistentPreRun(parent, args)
			}
			if parent.PersistentPreRunE != nil {
				err := parent.PersistentPreRunE(parent, args)
				if err != nil {
					return err
				}
			}
		}

		return originalPreRunE(cmd, args)
	}
	// cmd.PersistentFlags().AddFlagSet(config.GetKubeCtlFlagSet())
	originalRun := cmd.Run
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if err := kubectl.HandlePluginCommand(&kubectlPluginHandler{}, args); err != nil {
				return fmt.Errorf("kubectl plugin handler failed: %w", err)
			}
		}

		originalRun(cmd, args)
		return nil
	}

	logs.AddFlags(cmd.PersistentFlags())
	return cmd
}
