package version

import "runtime/debug"

// Version is set at build time with:
//
//	go build -ldflags "-X github.com/Tianbo-Qiu/whoopctl/internal/version.Version=v0.1.0"
var Version = ""

func String() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "dev"
}
