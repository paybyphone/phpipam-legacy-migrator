// Package session provides session management utility and token storage.
package session

import (
	"time"

	"github.com/imdario/mergo"
	"github.com/paybyphone/phpipam-sdk-go/phpipam"
)

// timeLayout represents the datetime format returned by the PHPIPAM api.
const timeLayout = "2006-01-02 15:04:05"

// Token represents a PHPIPAM session token
type Token struct {
	// The token string.
	String string `json:"token"`

	// The token's expiry date.
	Expires string
}

// Session represents a PHPIPAM session.
type Session struct {
	// The session's configuration.
	Config phpipam.Config

	// The session token.
	Token Token
}

// NewSession creates a new session based off supplied configs. It is up to the
// client for each controller implementation to log in and refresh the token.
// This is provided in the base client.Client implementation.
func NewSession(configs ...phpipam.Config) *Session {
	s := &Session{
		Config: phpipam.DefaultConfigProvider(),
	}
	for _, v := range configs {
		mergo.MergeWithOverwrite(&s.Config, v)
	}

	return s
}

// IsExpired checks to see if the token has expired via the date saved in
// SessionToken.
func (s *Session) IsExpired() bool {
	then, _ := time.Parse(timeLayout, s.Token.Expires)
	now := time.Now()

	return now.After(then)
}
