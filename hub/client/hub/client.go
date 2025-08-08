// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

// Package hub provides a client for interacting with the Agent Hub backend API, including agent management and related operations.
package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	corev1alpha1 "github.com/agntcy/dir/api/core/v1alpha1"
	v1alpha1 "github.com/agntcy/dir/hub/api/v1alpha1"
	"github.com/agntcy/dir/hub/sessionstore"
	"github.com/opencontainers/go-digest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const chunkSize = 4096 // 4KB

// Client defines the interface for interacting with the Agent Hub backend for agent operations.
type Client interface {
	// PushAgent uploads an agent to the hub and returns the response or an error.
	PushAgent(ctx context.Context, agent []byte, repository any) (*v1alpha1.PushRecordResponse, error)
	// PullAgent downloads an agent from the hub and returns the agent data or an error.
	PullAgent(ctx context.Context, request *v1alpha1.PullRecordRequest) ([]byte, error)
	// CreateAPIKey creates an API key for the specified role and returns the (clientId, secret) or an error.
	CreateAPIKey(ctx context.Context, session *sessionstore.HubSession, roleName string, organizationId string) (*v1alpha1.CreateApiKeyResponse, error)
	// DeleteAPIKey deletes an API key from the hub and returns the response or an error.
	DeleteAPIKey(ctx context.Context, session *sessionstore.HubSession, apikeyId string) (*v1alpha1.DeleteApiKeyResponse, error)
}

// client implements the Client interface for the Agent Hub backend.
type client struct {
	v1alpha1.AgentDirServiceClient
	v1alpha1.ApiKeyServiceClient
	v1alpha1.OrganizationServiceClient
	v1alpha1.UserServiceClient
}

// New creates a new Agent Hub client for the given server address.
// Returns the client or an error if the connection could not be established.
func New(serverAddr string) (*client, error) { //nolint:revive
	// Create connection
	conn, err := grpc.NewClient(
		serverAddr,
		//grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	return &client{
		AgentDirServiceClient: v1alpha1.NewAgentDirServiceClient(conn),
		ApiKeyServiceClient:   v1alpha1.NewApiKeyServiceClient(conn),
	}, nil
}

// PushAgent uploads an agent to the hub in chunks and returns the response or an error.
func (c *client) PushAgent(ctx context.Context, agent []byte, repository any) (*v1alpha1.PushRecordResponse, error) {
	fmt.Printf("###AXT:: PushAgent(): --> <--\n")
	var parsedAgent *corev1alpha1.Agent
	if err := json.Unmarshal(agent, &parsedAgent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent: %w", err)
	}
	fmt.Printf("###AXT:: PushAgent(): repository=%v\n", repository)
	fmt.Printf("###AXT:: PushAgent(): parsedAgent.Name=%v\n", parsedAgent.Name)

	d := digest.FromBytes(agent).String()
	t := corev1alpha1.ObjectType_OBJECT_TYPE_AGENT.String()

	ref := &corev1alpha1.ObjectRef{
		Digest:      d,
		Type:        t,
		Size:        uint64(len(agent)),
		Annotations: parsedAgent.GetAnnotations(),
	}

	stream, err := c.AgentDirServiceClient.PushRecord(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create push stream: %w", err)
	}

	buf := make([]byte, chunkSize)
	agentReader := bytes.NewReader(agent)

	for {
		var n int

		n, err = agentReader.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("failed to read data: %w", err)
		}

		if n == 0 {
			break
		}

		msg := &v1alpha1.PushRecordRequest{
			Model: &corev1alpha1.Object{
				Data: buf[:n],
				Ref:  ref,
			},
		}

		switch parsedRepo := repository.(type) {
		case *v1alpha1.PushAgentRequest_RepositoryName:
			msg.Repository = parsedRepo
			if parsedRepo.RepositoryName != parsedAgent.Name {
				return nil, fmt.Errorf("repository name mismatch: expected %s, got %s", parsedAgent.Name, parsedRepo.RepositoryName)
			}
			msg.OrganisationName, err = GetOrganisationNameFromRepository(parsedRepo.RepositoryName)
			if err != nil {
				return nil, fmt.Errorf("failed to parse organization name from repository: %w", err)
			}
		case *v1alpha1.PushAgentRequest_RepositoryId:
			msg.Repository = parsedRepo
			// In this case, we read the organization name from the agent
			msg.OrganisationName, err = GetOrganisationNameFromRepository(parsedAgent.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to parse organization name from agent: %w", err)
			}
		default:
			return nil, fmt.Errorf("unknown repository type: %T", repository)
		}

		if err = stream.Send(msg); err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("could not send object: %w", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("could not receive response: %w", err)
	}

	return resp, nil
}

func GetOrganisationNameFromRepository(repository string) (string, error) {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repository format: %s. Expected format is '<org>/<repo>'", repository)
	}
	return parts[0], nil
}

// PullAgent downloads an agent from the hub in chunks and returns the agent data or an error.
func (c *client) PullAgent(ctx context.Context, request *v1alpha1.PullRecordRequest) ([]byte, error) {
	stream, err := c.AgentDirServiceClient.PullRecord(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull stream: %w", err)
	}

	var buffer bytes.Buffer

	for {
		var chunk *v1alpha1.PullRecordResponse

		chunk, err = stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to receive chunk: %w", err)
		}

		buffer.Write(chunk.GetModel().GetData())
	}

	return buffer.Bytes(), nil
}

func (c *client) CreateAPIKey(ctx context.Context, session *sessionstore.HubSession, roleName string, organizationName string) (*v1alpha1.CreateApiKeyResponse, error) {
	roleValue, ok := v1alpha1.ProductRole_value[roleName]
	if !ok {
		return nil, fmt.Errorf("Unknown role: %w", roleValue)
	}
	req := &v1alpha1.CreateApiKeyRequest{
		Role:             v1alpha1.ProductRole(roleValue),
		OrganizationName: organizationName,
	}

	stream, err := c.ApiKeyServiceClient.CreateAPIKey(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	var chunk *v1alpha1.CreateApiKeyResponse

	chunk, err = stream.Recv()
	if errors.Is(err, io.EOF) {
		// Not an error
	} else if err != nil {
		return nil, fmt.Errorf("failed to receive chunk: %w", err)
	}

	return chunk, nil
}

func (c *client) DeleteAPIKey(ctx context.Context, session *sessionstore.HubSession, apikeyId string) (*v1alpha1.DeleteApiKeyResponse, error) {
	req := &v1alpha1.DeleteApiKeyRequest{
		Id: apikeyId,
	}

	resp, err := c.ApiKeyServiceClient.DeleteAPIKey(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete API key: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("received nil response from delete api key")
	}

	return resp, nil
}
