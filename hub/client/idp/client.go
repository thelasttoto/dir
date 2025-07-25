// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

// Package idp provides a client for interacting with the Identity Provider (IDP) API, including tenant and organization management.
package idp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	// Path to the IDP API for retreiving an AccessToken for ApiKey Access.
	AccessTokenEndpoint = "/v1/token"
	// Fixed Client ID for the access token request. THIS IS NOT THE CLIENT ID THAT WILL BE IN ACCESS TOKEN.
	AccessTokenClientID = "0oackfvbjvy65qVi41d7" // TO DO: Why does need to be fixed? Should be configurable?
	// Scope for the access token request.
	AccessTokenScope = "openid offline_access"
	// Grant type for the access token request.
	AccessTokenGrantType = "password"
)

// TenantListResponse represents a list of tenants returned by the IDP API.
type TenantListResponse struct {
	Tenants []*TenantResponse `json:"tenants"`
}

// TenantResponse represents a single tenant's information.
type TenantResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	OrganizationID string   `json:"organizationId"`
	Entitlements   []string `json:"entitlements"`
}

// GetTenantsInProductResponse contains the response from the GetTenantsInProduct API call.
type GetTenantsInProductResponse struct {
	TenantList *TenantListResponse
	Response   *http.Response
	Body       []byte
}

// GetAccessTokenResponse contains the response when requesting an access token from the IDP API.
type GetAccessTokenResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

// RequestModifier is a function that modifies an HTTP request before it is sent.
type RequestModifier func(*http.Request) error

// WithBearerToken returns a RequestModifier that adds a Bearer token to the Authorization header.
func WithBearerToken(token string) RequestModifier {
	return func(req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)

		return nil
	}
}

// Client defines the interface for interacting with the IDP API.
type Client interface {
	// GetTenantsInProduct retrieves the list of tenants for a given product ID.
	GetTenantsInProduct(ctx context.Context, productID string, modifier ...RequestModifier) (*GetTenantsInProductResponse, error)

	// GetAccessToken retrieves an access token using the provided username and secret.
	GetAccessTokenFromApiKey(ctx context.Context, username string, secret string) (*GetAccessTokenResponse, error)
}

// client implements the Client interface for the IDP API.
type client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new IDP client with the given base URL and HTTP client.
func NewClient(baseURL string, httpClient *http.Client) Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// GetTenantsInProduct retrieves the list of tenants for the specified product ID from the IDP APÒI.
// It applies any provided request modifiers (e.g., for authentication).
// Returns the response or an error if the request fails.
func (c *client) GetTenantsInProduct(ctx context.Context, productID string, modifiers ...RequestModifier) (*GetTenantsInProductResponse, error) {
	const path = "/v1alpha1/tenant"

	params := url.Values{}
	params.Set("product", productID)

	requestURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for _, m := range modifiers {
		if err = m(req); err != nil {
			return nil, fmt.Errorf("failed to modify request: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var body []byte
	if resp.Body != nil {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	}

	if resp.StatusCode == http.StatusOK {
		var parsedResponse *TenantListResponse
		if err = json.Unmarshal(body, &parsedResponse); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &GetTenantsInProductResponse{
			TenantList: parsedResponse,
			Response:   resp,
			Body:       body,
		}, nil
	}

	return &GetTenantsInProductResponse{
		TenantList: nil,
		Response:   resp,
		Body:       body,
	}, nil
}

func (c *client) GetAccessTokenFromApiKey(ctx context.Context, username string, secret string) (*GetAccessTokenResponse, error) {
	requestURL := fmt.Sprintf("%s%s", c.baseURL, AccessTokenEndpoint)

	data := url.Values{}
	data.Set("grant_type", AccessTokenGrantType)
	data.Set("username", username)
	data.Set("password", secret)
	data.Set("scope", AccessTokenScope)
	data.Set("client_id", AccessTokenClientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var body []byte
	if resp.Body != nil {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	}

	if resp.StatusCode == http.StatusOK {
		var parsedResponse *GetAccessTokenResponse
		if err = json.Unmarshal(body, &parsedResponse); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return parsedResponse, nil
	}

	return nil, fmt.Errorf("Bad response status code: %d, body: %s", resp.StatusCode, string(body))
}
