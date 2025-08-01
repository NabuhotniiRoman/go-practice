package main

import (
	"log"
	"os"

	"go-practice/internal/build"
	"go-practice/internal/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "OIDC API Server"
	app.Version = build.Version
	app.Usage = "OIDC API Server with configuration management"

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
