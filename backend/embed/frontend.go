package embed

import (
	"os"
)

// FrontendFilesExist checks if frontend files are available
func FrontendFilesExist() bool {
	_, err := os.Stat("static")
	return err == nil
}
