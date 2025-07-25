// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package objectsv2

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// LoadFromReader loads AgentRecord data from an io.Reader.
func (x *AgentRecord) LoadFromReader(reader io.Reader) ([]byte, error) {
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

// LoadFromFile loads AgentRecord data from a file.
func (x *AgentRecord) LoadFromFile(path string) ([]byte, error) {
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
