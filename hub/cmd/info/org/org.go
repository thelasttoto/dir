// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

// Package org provides the CLI commands for getting organization information.
package org

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	saasv1alpha1 "github.com/agntcy/dir/hub/api/v1alpha1"
	auth "github.com/agntcy/dir/hub/auth"
	authUtils "github.com/agntcy/dir/hub/auth/utils"
	hubClient "github.com/agntcy/dir/hub/client/hub"
	hubOptions "github.com/agntcy/dir/hub/cmd/options"
	"github.com/agntcy/dir/hub/sessionstore"
	"github.com/spf13/cobra"
)

// NewCommand creates the "org" subcommand for the info command.
// It gets information about a specific organization by ID.
// Returns the configured *cobra.Command.
func NewCommand(_ *hubOptions.HubOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org <organization-id>",
		Short: "Get information about a specific organization",
		Long: `Get detailed information about a specific organization by providing its ID.

This command retrieves and displays the organization details in JSON format.

Examples:
  dirctl hub info org 12345678-1234-1234-1234-123456789abc
  dirctl hub info org my-org-id`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.SetOut(os.Stdout)
				cmd.SetErr(os.Stderr)
				_ = cmd.Help()

				return nil
			}

			return nil
		},
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Retrieve session from context
		ctxSession := cmd.Context().Value(sessionstore.SessionContextKey)
		currentSession, ok := ctxSession.(*sessionstore.HubSession)

		if !ok || !auth.HasLoginCreds(currentSession) {
			return errors.New("no current session found. please login first")
		}

		hc, err := hubClient.New(currentSession.HubBackendAddress)
		if err != nil {
			return fmt.Errorf("failed to create hub client: %w", err)
		}

		ctx := authUtils.AddAuthToContext(cmd.Context(), currentSession)

		orgID := args[0]
		req := &saasv1alpha1.GetOrganizationRequest{
			Id: orgID,
		}

		org, err := hc.GetOrganization(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to get organization: %w", err)
		}

		renderOrganization(cmd.OutOrStdout(), org)

		return nil
	}

	return cmd
}

// renderOrganization renders organization information in a formatted table-like output.
func renderOrganization(stream io.Writer, orgWithRole *saasv1alpha1.OrganizationWithRole) {
	org := orgWithRole.GetOrganization()

	fmt.Fprintf(stream, "Organization Information:\n")
	fmt.Fprintf(stream, "  ID:          %s\n", org.GetId())
	fmt.Fprintf(stream, "  Name:        %s\n", org.GetName())

	if org.GetDescription() != "" {
		fmt.Fprintf(stream, "  Description: %s\n", org.GetDescription())
	}

	// Use raw role enum string (same as orgs list)
	fmt.Fprintf(stream, "  Role:        %s\n", orgWithRole.GetRole().String())

	if createdAt := org.GetCreatedAt(); createdAt != nil {
		createdTime := time.Unix(createdAt.GetSeconds(), int64(createdAt.GetNanos()))
		fmt.Fprintf(stream, "  Created:     %s\n", createdTime.Format(time.RFC3339))
	}

	if updatedAt := org.GetUpdatedAt(); updatedAt != nil {
		updatedTime := time.Unix(updatedAt.GetSeconds(), int64(updatedAt.GetNanos()))
		fmt.Fprintf(stream, "  Updated:     %s\n", updatedTime.Format(time.RFC3339))
	}
}
