// Package y is a dummy package used for copying into another package.
// The canonical import path should be stripped when it is copied.
package y // import "other.com/y"

// NOTHING HERE, but it has an external test that imports itself in y_test.go.
