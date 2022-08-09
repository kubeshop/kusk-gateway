package cmd

import "github.com/pterm/pterm"

const (
	trimSpace = 13
)

func trimWidth(msg string) string {
	if width := pterm.GetTerminalWidth(); width > trimSpace && len(msg)+trimSpace > width {
		msg = msg[:width-trimSpace]
	}
	return msg
}

func NewSpinner(text string) *pterm.SpinnerPrinter {
	pterm.Success.Prefix = pterm.Prefix{
		Text:  " ✔",
		Style: pterm.NewStyle(pterm.FgLightGreen),
	}

	pterm.Error.Prefix = pterm.Prefix{
		Text:  "    ERROR:",
		Style: pterm.NewStyle(pterm.BgLightRed, pterm.FgBlack),
	}

	pterm.Info.Prefix = pterm.Prefix{
		Text: " •",
	}

	spinner, err := pterm.DefaultSpinner.
		WithRemoveWhenDone(false).
		// Src: https://github.com/gernest/wow/blob/master/spin/spinners.go#L335
		WithSequence(`  ⠋ `, `  ⠙ `, `  ⠹ `, `  ⠸ `, `  ⠼ `, `  ⠴ `, `  ⠦ `, `  ⠧ `, `  ⠇ `, `  ⠏ `).
		Start(trimWidth(text))

	if err != nil {
		panic(err)
	}

	return spinner
}
