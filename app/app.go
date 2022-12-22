package app

var (
	AppVersion = "v0.0.1-pre-alpha.1" // this can be injected at build time with something like: go build -ldflags "-X github.com/pokt-network/pocket/app.AppVersion=v1.0.1" ./app/pocket
	CommitHash = "unknown"            // this can be injected at build time with something like: go build -ldflags "-X github.com/pokt-network/pocket/app.CommitHash=$(git rev-parse HEAD)" ./app/pocket
	BuildDate  = "unknown"            // this can be injected at build time with something like: go build -ldflags "-X github.com/pokt-network/pocket/app.BuildDate=$(date -u '+%Y-%m-%d_%I:%M:%S%p')" ./app/pocket
)
