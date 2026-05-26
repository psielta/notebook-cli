package version

// Version is replaced by release builds with:
//
//	go build -ldflags "-X notebook-cli/internal/version.Version=v1.2.3"
var Version = "dev"
