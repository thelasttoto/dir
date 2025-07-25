// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"errors"
	"fmt"
	"io"
	"os"
)

func GetReader(fpath string, fromStdin bool) (io.ReadCloser, error) {
	if fpath == "" && !fromStdin {
		return nil, errors.New("if no path defined --stdin flag must be set")
	}

	if fpath != "" {
		file, err := os.Open(fpath)
		if err != nil {
			return nil, fmt.Errorf("could not open file %s: %w", fpath, err)
		}

		return file, nil
	}

	return os.Stdin, nil
}
