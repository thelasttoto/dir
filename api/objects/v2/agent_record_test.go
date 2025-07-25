// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package objectsv2

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentRecord_LoadFromReader(t *testing.T) {
	agentRecord := &AgentRecord{}
	data := `{"name": "TestAgent", "version": "1.0", "schema_version": "v0.4.0"}`
	reader := bytes.NewReader([]byte(data))

	_, err := agentRecord.LoadFromReader(reader)
	require.NoError(t, err)

	assert.Equal(t, "TestAgent", agentRecord.GetName())
	assert.Equal(t, "1.0", agentRecord.GetVersion())
	assert.Equal(t, "v0.4.0", agentRecord.GetSchemaVersion())
}

func TestAgentRecord_LoadFromReader_InvalidJSON(t *testing.T) {
	agentRecord := &AgentRecord{}
	data := `{"name": "TestAgent", "version":`
	reader := bytes.NewReader([]byte(data))

	_, err := agentRecord.LoadFromReader(reader)
	assert.Error(t, err) //nolint:testifylint // Test should fail
	assert.Contains(t, err.Error(), "failed to unmarshal data")
}
