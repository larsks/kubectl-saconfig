package version

import (
	"fmt"
	"runtime/debug"

	"github.com/larsks/gobot/tools"
)

var (
	Version = "development"
)

func ShowVersion() {
	bi, ok := debug.ReadBuildInfo()
	fmt.Printf("kubectl-saconfig version %s", Version)
	if ok {
		bimap := tools.BuildInfoMap(bi)
		fmt.Printf(" %s/%s", bimap["GOOS"], bimap["GOARCH"])
		if vcs, ok := bimap["vcs"]; ok && vcs == "git" {
			fmt.Printf(" revision %s on %s", bimap["vcs.revision"][:10], bimap["vcs.time"])
		}
	}
	fmt.Printf("\n")
}
