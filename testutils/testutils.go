// Package testutils implements helper functions for running unittests.
package testutils

import (
	"io"
	"os"
)

// IsDirEmpty is an internal helper method for determining if a directory has
// files.
func IsDirEmpty(name string) (bool, error) {
	fd, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer fd.Close()
	_, err = fd.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
