// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package objectsv3

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// LoadFromReader loads Record data from an io.Reader.
func (x *Record) LoadFromReader(reader io.Reader) ([]byte, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	err = json.Unmarshal(data, x)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return data, nil
}

// LoadFromFile loads Record data from a file.
func (x *Record) LoadFromFile(path string) ([]byte, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer reader.Close()

	data, err := x.LoadFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to load from reader: %w", err)
	}

	return data, nil
}
