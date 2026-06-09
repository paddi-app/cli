package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows requires the empty string "" as the first parameter for the title argument sometimes
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		// macOS
		cmd = exec.Command("open", url)
	case "linux":
		// Linux
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
