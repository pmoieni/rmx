package google

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pmoieni/rmx/internal/lib"
	"golang.org/x/oauth2"
)

var (
	stateLength   uint = 16
	defaultScopes      = []string{"profile", "email"}
)

type Provider struct {
	*oauth2.Config

	oidc *oidc.Provider
}

func NewOIDC() *Provider {
	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		log.Fatal(err)
	}

	return &Provider{
		Config: &oauth2.Config{
			Endpoint: provider.Endpoint(),
			Scopes:   append([]string{oidc.ScopeOpenID}, defaultScopes...),
		},
		oidc: provider,
	}
}

func (gp *Provider) GetClientID() string {
	return gp.ClientID
}

func (gp *Provider) GetClienSecret() string {
	return gp.ClientSecret
}

func (gp *Provider) GetRedirectURL() string {
	return gp.RedirectURL
}

func (gp *Provider) GetScopes() []string {
	return gp.Scopes
}

func (gp *Provider) SetClientID(clientID string) {
	gp.ClientID = clientID
}

func (gp *Provider) SetClientSecret(clientSecret string) {
	gp.ClientSecret = clientSecret
}

func (gp *Provider) SetRedirectURL(redirectURL string) {
	gp.RedirectURL = redirectURL
}

func (gp *Provider) SetScopes(scopes []string) {
	gp.Scopes = append(gp.Scopes, scopes...)
}

func (gp *Provider) HandleAuthorizationRequest(w http.ResponseWriter, r *http.Request) {
	state, err := lib.RandomString(stateLength)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, gp.AuthCodeURL(state, oauth2.AccessTypeOffline), http.StatusTemporaryRedirect)
}

func (gp *Provider) HandleCallbackRequest(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	token, err := gp.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userInfo, err := gp.oidc.UserInfo(r.Context(), oauth2.StaticTokenSource(token))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := struct {
		Token    *oauth2.Token  `json:"token"`
		UserInfo *oidc.UserInfo `json:"userInfo"`
	}{token, userInfo}
	data, err := json.MarshalIndent(resp, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(data)
}
