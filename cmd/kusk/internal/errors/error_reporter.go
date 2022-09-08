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
	cmd      *cobra.Command
	err      error
	miscInfo map[string]interface{}
}

// NewErrorReporter ...
func NewErrorReporter(cmd *cobra.Command, err error, miscInfo map[string]interface{}) ErrorReporter {
	return &errorReporter{
		cmd:      cmd,
		err:      err,
		miscInfo: miscInfo,
	}
}

// ReportError ...
func (r *errorReporter) Report() {
	commandName := r.cmd.Use
	if debug {
		err := strings.ReplaceAll(r.err.Error(), "\n", " - ")
		fmt.Fprintf(os.Stderr, "kusk: reporting error - command=%q, err=%q, miscInfo=%v\n", commandName, err, r.miscInfo)
	}
	analytics.SendAnonymousCommandError(commandName, r.err, r.miscInfo)
}
