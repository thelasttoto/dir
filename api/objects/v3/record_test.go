// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package objectsv3

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecord_LoadFromReader(t *testing.T) {
	record := &Record{}
	data := `{"name": "TestRecord", "version": "1.0", "schema_version": "v0.5.0"}`
	reader := bytes.NewReader([]byte(data))

	_, err := record.LoadFromReader(reader)
	require.NoError(t, err)

	assert.Equal(t, "TestRecord", record.GetName())
	assert.Equal(t, "1.0", record.GetVersion())
	assert.Equal(t, "v0.5.0", record.GetSchemaVersion())
}

func TestRecord_LoadFromReader_InvalidJSON(t *testing.T) {
	record := &Record{}
	data := `{"name": "TestRecord", "version":`
	reader := bytes.NewReader([]byte(data))

	_, err := record.LoadFromReader(reader)
	assert.Error(t, err) //nolint:testifylint // Test should fail
	assert.Contains(t, err.Error(), "failed to unmarshal data")
}
