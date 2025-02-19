package config

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func SetupGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     Config("GOOGLE_CLIENT_ID"),
		ClientSecret: Config("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  Config("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}
