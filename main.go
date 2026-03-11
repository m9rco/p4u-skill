package main

import "github.com/m9rco/p4u-skill/cmd"

// version is set at build time via -ldflags "-X main.version=<value>".
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
