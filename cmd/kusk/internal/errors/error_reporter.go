package errors

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kubeshop/kusk-gateway/pkg/analytics"
)

var (
	debug = false
)

// ErrorReporter ...
type ErrorReporter interface {
	Report()
}

type errorReporter struct {
	cmd *cobra.Command
	err error
}

// NewErrorReporter ...
func NewErrorReporter(cmd *cobra.Command, err error) ErrorReporter {
	return &errorReporter{
		cmd: cmd,
		err: err,
	}
}

// ReportError ...
func (r *errorReporter) Report() {
	commandName := r.cmd.Use
	if debug {
		err := strings.ReplaceAll(r.err.Error(), "\n", " - ")
		fmt.Fprintf(os.Stderr, "analytics: reporting error - command=%v, err=%v\n", commandName, err)
	}

	err := analytics.SendAnonymousCMDInfo(r.err)

	if debug && err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("analytics: report error failed - command=%v, err=%v, %w", commandName, r.err, err))
	}
}
