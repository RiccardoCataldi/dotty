package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func copyPathToClipboard(path string) (string, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return "", fmt.Errorf("no clipboard tool found (install xclip or xsel)")
		}
	default:
		return "", fmt.Errorf("clipboard not supported on this platform")
	}

	cmd.Stdin = strings.NewReader(path)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return "Copied to clipboard", nil
}
