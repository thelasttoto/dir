// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"

	v1alpha1 "github.com/agntcy/dir/hub/api/v1alpha1"
	"github.com/agntcy/dir/hub/auth"
	hubClient "github.com/agntcy/dir/hub/client/hub"
	"github.com/agntcy/dir/hub/sessionstore"
)

// CreateAPIKey creates a new API key in the hub and returns the response.
// It uses the provided session for authentication.
func CreateAPIKey(
	ctx context.Context,
	hc hubClient.Client,
	role string,
	organizationId string,
	organizationName string,
	session *sessionstore.HubSession,
) (*v1alpha1.CreateApiKeyResponse, error) {
	ctx = auth.AddAuthToContext(ctx, session)

	resp, err := hc.CreateAPIKey(ctx, nil, role, organizationId, organizationName)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return resp, nil
}

// DeleteAPIKey deletes an API key from the hub and returns the response.
// It uses the provided session for authentication.
func DeleteAPIKey(
	ctx context.Context,
	hc hubClient.Client,
	apikeyId string,
	session *sessionstore.HubSession,
) (*v1alpha1.DeleteApiKeyResponse, error) {
	ctx = auth.AddAuthToContext(ctx, session)

	resp, err := hc.DeleteAPIKey(ctx, nil, apikeyId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete API key: %w", err)
	}

	return resp, nil
}
