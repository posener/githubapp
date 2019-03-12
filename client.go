// Package githubapp provides oauth2 Github app authentication client.
//
// According to https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps.
//
// Usage
//
// 	func main() {
// 		ctx := context.Background()
// 		cfg := githubapp.Config{
//			AppID: "1234",
//			PrivateKey: []byte(os.Getenv("GITHUB_APP_PRIVATE_KEY"))
//		}
// 		c := cfg.Client(ctx)
// 		// Use c...
// 	}
//
// The created client can be used to create a github API client
// with the github.com/google/go-github/github library.
// Once your application will have installation, you would like to
// get application clients.
//
// 		app := cfg.NewApp(ctx)
// 		installation, err := app.Installation(ctx, "<github-login>")
// 		// Check err and use installation...
//
// The installation has an authenticated http client and github API client
// ready to be used.
package githubapp

import (
	"context"
	"crypto/rsa"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
)

// maxExpires is the maximum expiration time of github app token,
// as defined by github.
const maxExpires = 10 * time.Minute

var defaultHeader = jws.Header{Algorithm: "RS256", Typ: "JWT"}

// Config of Application authentication.
type Config struct {
	// AppID is the application ID from github app settings.
	AppID string
	// PrivateKey is the bytes of the private key from github app
	// setting. It can be used by reading the file content.
	// Usually it will be preferred to use it from environment variable.
	// The value of environment variable GITHUB_PRIVATE_KEY of file
	// 'private_key.pem' could be set by:
	//
	// 	export GITHUB_PRIVATE_KEY="$(cat private_key.pem)"
	//
	// Then, it is possible to use this environment variable with this
	// field by `PrivateKey: []byte(os.Getenv("GITHUB_PRIVATE_KEY"))`.
	PrivateKey []byte
	// expire is the duration that the app token expire. 10 minutes
	// is the maximal value.
	Expire time.Duration
}

// Client returns an http.Client with oauth2 transport that authenticate
// request using an application signed JWT token.
// Client will panic if the give private key is invalid.
func (c *Config) Client(ctx context.Context) *http.Client {
	return oauth2.NewClient(ctx, c.TokenSource())
}

// TokenSource returns a token source for github application.
// TokenSource will panic if the give private key is invalid.
func (c *Config) TokenSource() oauth2.TokenSource {
	pk, err := jwt.ParseRSAPrivateKeyFromPEM(c.PrivateKey)
	if err != nil {
		panic(err)
	}
	return oauth2.ReuseTokenSource(nil, appSource{
		appID:  c.AppID,
		expire: c.Expire,
		pk:     pk,
	})
}

type appSource struct {
	appID  string
	expire time.Duration
	pk     *rsa.PrivateKey
}

func (js appSource) Token() (*oauth2.Token, error) {
	// Adjust expire duration to the maximum allowed.
	if js.expire <= 0 || js.expire > maxExpires {
		js.expire = maxExpires
	}
	exp := time.Now().Add(js.expire)
	claimSet := &jws.ClaimSet{Iss: js.appID, Exp: exp.Unix()}
	h := defaultHeader
	payload, err := jws.Encode(&h, claimSet, js.pk)
	if err != nil {
		return nil, err
	}
	return &oauth2.Token{TokenType: "bearer", AccessToken: payload, Expiry: exp}, nil
}
