package githubapp

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

// Cache is a common interface for cache.
type Cache interface {
	// Get returns an object from the cache by key. If the object does
	// not exists, it should return nil.
	Get(string) interface{}
	// Set sets an object by a string key in the cache.
	Set(string, interface{})
}

// App is a struct for github application that can produce installation
// clients.
type App struct {
	// Client is github API with the App credentials.
	*github.Client

	cfg   Config
	cache Cache
	mu    sync.RWMutex
}

// Installation holds installation clients and information.
type Installation struct {
	// Client is an http client with the installation credentials.
	*http.Client
	// Github is a Github API client with the installation credentials.
	Github *github.Client
	// ID is the installation ID.
	ID int
}

// Option is an option for new applications.
type Option func(*App)

// OptWithCache is an option to use cache to hold the clients.
func OptWithCache(c Cache) Option {
	return func(a *App) {
		a.cache = c
	}
}

// NewApp returns a Github app object.
func (c *Config) NewApp(ctx context.Context, opts ...Option) *App {
	a := &App{
		Client: github.NewClient(c.Client(ctx)),
		cfg:    *c,
	}

	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Installation returns github installation client for a given user login.
func (a *App) Installation(ctx context.Context, login string) (*Installation, error) {
	inst := a.fromCache(login)
	if inst != nil {
		return inst, nil
	}

	install, _, err := a.Client.Apps.FindUserInstallation(ctx, login)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting user installation")
	}

	installID := int(install.GetID())

	appID, _ := strconv.Atoi(a.cfg.AppID)
	tr, err := ghinstallation.New(http.DefaultTransport, appID, installID, a.cfg.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "get install transport")
	}
	cl := &http.Client{Transport: tr}
	inst = &Installation{
		Client: cl,
		Github: github.NewClient(cl),
		ID:     installID,
	}
	a.toCache(login, inst)
	return inst, nil
}

func (a *App) fromCache(login string) *Installation {
	if a.cache == nil {
		return nil
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	i := a.cache.Get(cacheKey(login))
	if i == nil {
		return nil
	}
	return i.(*Installation)
}

func (a *App) toCache(login string, i *Installation) {
	if a.cache == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cache.Set(cacheKey(login), i)
}

func cacheKey(login string) string {
	return "installation/" + login
}
