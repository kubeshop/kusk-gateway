// MIT License
//
// Copyright (c) 2022 Kubeshop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package utils

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

// NewSpinner ...
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
