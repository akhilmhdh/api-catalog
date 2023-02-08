package main

import "github.com/1-platform/api-catalog/internal/cli"

var version = "development"

func main() {
	cli.Run(version)
}
