package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pmoieni/rmx/internal/lib"
	"github.com/pmoieni/rmx/internal/oauth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	stateLength   uint = 16
	defaultScopes      = []string{"read:user", "user:email"}

	ProfileURL = "https://api.github.com/user"
	EmailURL   = "https://api.github.com/user/emails"
)

type Provider struct {
	*oauth2.Config
}

func NewOAuth2(ctx context.Context, clientID, clientSecret, redirectURL string) *Provider {
	return &Provider{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     github.Endpoint,
			Scopes:       defaultScopes,
		},
	}
}

func (p *Provider) Name() string { return "github" }

func (p *Provider) HandleAuthorizationRequest(w http.ResponseWriter, r *http.Request) {
	state, err := lib.RandomString(stateLength)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	setCallbackCookie(w, "state", state)

	http.Redirect(w, r, p.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func setCallbackCookie(w http.ResponseWriter, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, c)
}

// TODO: implement a way to avoid fetching new token if previous token is still valid
func (p *Provider) GetCallbackResult(r *http.Request) (*oauth.CallbackResult, error) {
	oauthToken, err := p.getOAuthToken(r)
	if err != nil {
		return nil, err
	}

	return p.FetchUser(r, oauthToken)
}

func (p *Provider) getOAuthToken(r *http.Request) (*oauth2.Token, error) {
	state, err := r.Cookie("state")
	if err != nil {
		return nil, errors.New("state not found")
	}

	originalState := r.URL.Query().Get("state")

	if originalState != "" && (originalState != state.Value) {
		return nil, errors.New("state didn't match")
	}

	token, err := p.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		return nil, errors.New("authorization failed")
	}

	if !token.Valid() {
		return nil, errors.New("invalid token")
	}

	return token, err
}

func (p *Provider) FetchUser(r *http.Request, token *oauth2.Token) (*oauth.CallbackResult, error) {
	req, err := http.NewRequest("GET", ProfileURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	// TODO: is it ok to use default client here?
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API responded with a %d trying to fetch user information", response.StatusCode)
	}

	bs, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	res := &oauth.CallbackResult{}
	res.Issuer = p.Name()

	if err := json.NewDecoder(bytes.NewReader(bs)).Decode(&res.RawData); err != nil {
		return nil, err
	}

	if err := userFromReader(bytes.NewReader(bs), res); err != nil {
		return nil, err
	}

	if res.Email == "" {
		for _, scope := range defaultScopes {
			if strings.TrimSpace(scope) == "user" || strings.TrimSpace(scope) == "user:email" {
				res.Email, err = getPrivateMail(token)
				if err != nil {
					return nil, err
				}
				break
			}
		}
	}

	return res, nil
}

func userFromReader(reader io.Reader, res *oauth.CallbackResult) error {
	u := struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}{}

	err := json.NewDecoder(reader).Decode(&u)
	if err != nil {
		return err
	}

	res.UserID = strconv.Itoa(u.ID)
	res.Email = u.Email

	return err
}

func getPrivateMail(token *oauth2.Token) (string, error) {
	req, err := http.NewRequest("GET", EmailURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	// TODO: is it ok to use default client here?
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		if response != nil {
			// TODO: log these errors
			response.Body.Close()
		}
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API responded with a %d trying to fetch user email", response.StatusCode)
	}

	var mailList []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(response.Body).Decode(&mailList); err != nil {
		return "", err
	}

	for _, v := range mailList {
		if v.Primary && v.Verified {
			return v.Email, nil
		}
	}

	return "", errors.New("no verified GitHub email found")
}
