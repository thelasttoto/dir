// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

// Package orgs provides the CLI command for listing organizations (organizations) for the logged-in user.
// The orgs command only lists organizations/organizations. To switch between organizations, use the "orgs switch" subcommand.
package orgs

import (
	"errors"
	"fmt"
	"io"

	saasv1alpha1 "github.com/agntcy/dir/hub/api/v1alpha1"
	auth "github.com/agntcy/dir/hub/auth"
	hubClient "github.com/agntcy/dir/hub/client/hub"
	hubOptions "github.com/agntcy/dir/hub/cmd/options"
	"github.com/agntcy/dir/hub/sessionstore"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

const (
	nameHeader = "Organization Name"
	idHeader   = "Organization ID"
	roleHeader = "Role"
	gapSize    = 4
)

// NewCommand creates the "orgs" command for the Agent Hub CLI.
// It lists organizations (organizations) for the logged-in user.
// Returns the configured *cobra.Command.
func NewCommand(hubOpts *hubOptions.HubOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "orgs",
		Aliases: []string{"organizations"},
		Short:   "List organizations for logged in user",
	}

	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
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

		ctx := auth.AddAuthToContext(cmd.Context(), currentSession)

		orgs, err := hc.ListOrganizations(ctx, &saasv1alpha1.ListOrganizationsRequest{})
		if err != nil {
			return fmt.Errorf("failed to get orgs list: %w", err)
		}

		renderList(cmd.OutOrStdout(), orgs.Organizations)

		return nil
	}

	return cmd
}

type renderFn func(int, int, int) string

func renderList(stream io.Writer, organizationsWithRoles []*saasv1alpha1.OrganizationWithRole) {
	renderFns := make([]renderFn, len(organizationsWithRoles))

	longestNameLen := len(nameHeader) // Start with header length
	longestRoleLen := len(roleHeader) // Start with header length
	longestIDLen := len(idHeader)     // Start with header length

	for i, org := range organizationsWithRoles {
		if len(org.Organization.Name) > longestNameLen {
			longestNameLen = len(org.Organization.Name)
		}

		if len(org.Organization.Id) > longestIDLen {
			longestIDLen = len(org.Organization.Id)
		}

		if len(org.Role.String()) > longestRoleLen {
			longestRoleLen = len(org.Role.String())
		}

		renderFns[i] = func(lName, lId, lRole int) string {

			nameCol := text.AlignLeft.Apply(org.Organization.Name, lName+gapSize)
			idCol := text.AlignLeft.Apply(org.Organization.Id, lId+gapSize)
			roleCol := text.AlignLeft.Apply(org.Role.String(), lRole)

			return fmt.Sprintf("%s%s%s", nameCol, idCol, roleCol)
		}
	}

	// Print headers
	nameHeader := text.AlignLeft.Apply(nameHeader, longestNameLen+gapSize)
	idHeader := text.AlignLeft.Apply(idHeader, longestIDLen+gapSize)
	roleHeader := text.AlignLeft.Apply(roleHeader, longestRoleLen)
	fmt.Fprintf(stream, "%s%s%s\n", nameHeader, idHeader, roleHeader) //nolint:errcheck

	for _, organization := range renderFns {
		fmt.Fprintln(stream, organization(longestNameLen, longestIDLen, longestRoleLen)) //nolint:errcheck
	}
}
