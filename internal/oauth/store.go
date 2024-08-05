package oauth

import (
	"net/http"
	"sync"
)

type Provider interface {
	GetClientID() string
	GetClientSecret() string
	GetRedirectURL() string
	GetScope() string

	SetClientID(string)
	SetClientSecret(string)
	SetRedirectURL(string)
	SetScope(string)

	HandleAuthorizationRequest() http.HandlerFunc
	HandleCallbackRequest() http.HandlerFunc
}

type ClientStore struct {
	sync.RWMutex

	clients map[string]Provider
}
