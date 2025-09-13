package main

import (
	"flag"
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/version"
)

func main() {
	ver := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *ver {
		fmt.Printf("hottgo %s (%s, %s)\n", version.Version, version.Commit, version.Date)
		return
	}
	// TODO: CLI real entrypoint
	fmt.Println("hottgo: kernel playground (use -version)")
}
