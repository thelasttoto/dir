// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"github.com/agntcy/dir/hub/cmd/options"
	"github.com/spf13/cobra"
)

type ApiKeyCreateOptions struct {
	*options.HubOptions

	Role string
}

func NewApiKeyCreateOptions(hubOpts *options.HubOptions, cmd *cobra.Command) *ApiKeyCreateOptions {
	opt := &ApiKeyCreateOptions{
		HubOptions: hubOpts,
	}

	opt.AddRegisterFn(func() error {
		cmd.Flags().StringVarP(&opt.Role, "role", "r", "", "Role name")

		return nil
	})

	return opt
}
