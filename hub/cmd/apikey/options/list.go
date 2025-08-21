// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"github.com/agntcy/dir/hub/cmd/options"
	"github.com/spf13/cobra"
)

type ApiKeyListOptions struct {
	*options.HubOptions

	OrganizationId string
}

func NewApiKeyListOptions(hubOpts *options.HubOptions, cmd *cobra.Command) *ApiKeyListOptions {
	opt := &ApiKeyListOptions{
		HubOptions: hubOpts,
	}

	opt.AddRegisterFn(func() error {
		cmd.Flags().StringVarP(&opt.OrganizationId, "org-id", "o", "", "Organization ID")

		return nil
	})

	return opt
}
