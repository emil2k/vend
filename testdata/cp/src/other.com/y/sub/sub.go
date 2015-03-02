// Package sub is a dummy package used for copying into another package.
// The canonical import path should be stripped when it is copied.
// Added some extra whitespace to the inline comment to make sure it is
// stripped properly.
package sub        		 //       import "other.com/y/sub"
const TC int = 1

// NOTHING HERE, but it has an external test that imports itself in y_test.go.
