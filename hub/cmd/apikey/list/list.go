// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/agntcy/dir/hub/auth"
	hubClient "github.com/agntcy/dir/hub/client/hub"
	"github.com/agntcy/dir/hub/cmd/apikey/options"
	hubOptions "github.com/agntcy/dir/hub/cmd/options"
	service "github.com/agntcy/dir/hub/service"
	"github.com/agntcy/dir/hub/sessionstore"
	"github.com/spf13/cobra"
)

// NewCommand creates the "list" command for the Agent apikey Hub CLI.
func NewCommand(hubOpts *hubOptions.HubOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list --org-id <org_id>",
		Short: "list API keys for a specific organization",
		Long: `list API keys for a specific organization.

Parameters:
  --org-id <org_id>  Organization ID

Example:
  # List API keys for organization "MyOrg"
  dirctl hub apikey list --org-id MyOrg`,
	}

	opts := options.NewApiKeyListOptions(hubOpts, cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runCommand(cmd, args, opts)
	}

	return cmd
}

func runCommand(cmd *cobra.Command, args []string, opts *options.ApiKeyListOptions) error {
	if opts.OrganizationId == "" {
		return errors.New("organization ID is required")
	}

	// Retrieve session from context
	ctxSession := cmd.Context().Value(sessionstore.SessionContextKey)
	currentSession, ok := ctxSession.(*sessionstore.HubSession)
	if !ok || currentSession == nil {
		return errors.New("could not get current hub session")
	}
	if !auth.HasLoginCreds(currentSession) && auth.HasApiKey(currentSession) {
		fmt.Println("User is authenticated with API key, using it to get credentials...")
		if err := auth.RefreshApiKeyAccessToken(cmd.Context(), currentSession, opts.ServerAddress); err != nil {
			return fmt.Errorf("failed to refresh API key access token: %w", err)
		}
	}

	if !auth.HasLoginCreds(currentSession) && !auth.HasApiKey(currentSession) {
		return errors.New("you need to be logged in to push to the hub\nuse `dirctl hub login` command to login")
	}

	hc, err := hubClient.New(currentSession.HubBackendAddress)
	if err != nil {
		return fmt.Errorf("failed to create hub client: %w", err)
	}

	resp, err := service.ListAPIKeys(cmd.Context(), hc, opts.OrganizationId, currentSession)
	if err != nil {
		return fmt.Errorf("failed to list API keys: %w", err)
	}

	fmt.Printf("API Keys for organization %s:\n", opts.OrganizationId)
	prettyModel, err := json.MarshalIndent(resp.Apikeys, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal agent: %w", err)
	}
	fmt.Fprintf(os.Stdout, "%s\n", string(prettyModel))

	return nil
}
