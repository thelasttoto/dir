// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package create

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

// NewCommand creates the "create" command for the Agent apikey Hub CLI.
func NewCommand(hubOpts *hubOptions.HubOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a new API key for hub actions",
		Long: `create a new API key for hub actions.

Example:
  # Create a new API key
  dirctl hub apikey create --role PRODUCT_ROLE_ADMIN`,
	}

	opts := options.NewApiKeyCreateOptions(hubOpts, cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runCommand(cmd, args, opts)
	}

	return cmd
}

func runCommand(cmd *cobra.Command, args []string, opts *options.ApiKeyCreateOptions) error {
	fmt.Printf("Creating new API key with role: '%s' for organization: '%s'\n", opts.Role, opts.OrganizationId)

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

	resp, err := service.CreateAPIKey(cmd.Context(), hc, opts.Role, opts.OrganizationId, currentSession)
	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	fmt.Printf("API Key created successfully: %s\n", resp)

	prettyModel, err := json.MarshalIndent(resp.Token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal agent: %w", err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", string(prettyModel))
	return nil
}
