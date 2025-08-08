// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"fmt"
	"net/url"

	"github.com/agntcy/dir/hub/config"
	"github.com/agntcy/dir/hub/sessionstore"
	"github.com/pkg/browser"
)

const (
	loginPath string = "login"
)

// OpenBrowserForLogin opens the default web browser to the login page for the given AuthConfig.
func OpenBrowserForLogin(authConfig *sessionstore.AuthConfig) error {
	params := url.Values{}
	params.Add("redirectUri", fmt.Sprintf("http://localhost:%d", config.LocalWebserverPort))

	loginPageWithRedirect := fmt.Sprintf("%s/%s/%s?%s", authConfig.IdpFrontendAddress, authConfig.IdpProductID, loginPath, params.Encode())

	return browser.OpenURL(loginPageWithRedirect) //nolint:wrapcheck
}
