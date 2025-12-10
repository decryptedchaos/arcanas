/*
 * This file is part of the Arcanas project.
 * Licensed under the Mozilla Public License, v. 2.0.
 * See the LICENSE file at the project root for details.
 */

package embed

import (
	"os"
)

// FrontendFilesExist checks if frontend files are available
func FrontendFilesExist() bool {
	_, err := os.Stat("static")
	return err == nil
}
