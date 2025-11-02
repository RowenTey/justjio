package config

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func SetupGoogleOAuthConfig(conf *Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     conf.GoogleOauth.ClientID,
		ClientSecret: conf.GoogleOauth.ClientSecret,
		RedirectURL:  conf.GoogleOauth.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}
