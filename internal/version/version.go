package version

// Version is the binary version, injected at build time via:
//
//	-ldflags "-X github.com/codref/stdix/internal/version.Version=<tag>"
//
// Falls back to "dev" when built without ldflags (e.g. go run or go build without Makefile).
var Version = "dev"
