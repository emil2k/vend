// Package y_test is an external test to test copying of a package into another
// package.

package y_test

import (
	_ "other.com/y" // this should update when moved into a new location.
)
