package version

import (
	"fmt"
	"runtime"
)

var (
	Version = "v0.2.0"
)

func Print() {
	fmt.Printf("aro-rp-versions\nversion %v\ngo version: %v\n",
		Version, runtime.Version())
}
