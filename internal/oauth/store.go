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

	HandleAuthorizationRequest(http.ResponseWriter, *http.Request)
	HandleCallbackRequest(http.ResponseWriter, *http.Request)
}

type ClientStore struct {
	sync.RWMutex

	clients map[string]Provider
}

func NewClientStore() *ClientStore {
	return &ClientStore{
		clients: make(map[string]Provider),
	}
}

func (cs *ClientStore) AddProvider(name string, provider Provider) {
	cs.Lock()
	defer cs.Unlock()

	cs.clients[name] = provider
}

func (cs *ClientStore) GetProvider(name string) Provider {
	cs.RLock()
	defer cs.RUnlock()

	return cs.clients[name]
}
