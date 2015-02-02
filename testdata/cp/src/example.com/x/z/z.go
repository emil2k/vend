// Package z is used for testing recursive cp commands it's import should only
// update when the cp command is intended to recursively update the import paths.
package z

import (
	_ "other.com/y" // this path should be modified
)
