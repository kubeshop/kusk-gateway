package kuskui

import (
	"fmt"
	"strings"

	"github.com/gookit/color"
)

func PrintWarning(messages ...string) {
	fmt.Println(color.FgLightYellow.Render("❕ " + strings.Join(messages, ", ")))
}

func PrintSuccess(messages ...string) {
	fmt.Println(color.FgLightGreen.Render("🎉 " + strings.Join(messages, ", ")))
}

func PrintError(messages ...string) {
	fmt.Println(color.FgRed.Render("❌ " + strings.Join(messages, ", ")))
}

func PrintStart(messages ...string) {
	fmt.Println(color.FgWhite.Render("🚢 " + strings.Join(messages, ", ")))
}

func PrintInfo(messages ...string) {
	fmt.Println(color.FgWhite.Render(strings.Join(messages, ", ")))
}
