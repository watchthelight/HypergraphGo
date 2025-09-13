package main

import (
	"flag"
	"fmt"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	ver := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *ver {
		fmt.Printf("hottgo %s (%s, %s)\n", version, commit, date)
		return
	}
	// TODO: CLI real entrypoint
	fmt.Println("hottgo: kernel playground (use -version)")
}
