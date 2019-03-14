# githubapp

Package githubapp provides oauth2 Github app authentication client.

According to [https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps](https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps).

Usage

```go
func main() {
	ctx := context.Background()
	cfg := githubapp.Config{
		AppID: "1234",
		PrivateKey: []byte(os.Getenv("GITHUB_APP_PRIVATE_KEY"))
	}
	c := cfg.Client(ctx)
	// Use c...
}
```

The created client can be used to create a github API client
with the github.com/google/go-github/github library.
Once your application will have installation, you would like to
get application clients.

```go
app := cfg.NewApp(ctx)
installation, err := app.Installation(ctx, "<github-login>")
// Check err and use installation...
```

The installation has an authenticated http client and github API client
ready to be used.

## Sub Packages

* [cache](./cache)


---

Created by [goreadme](https://github.com/apps/goreadme)
