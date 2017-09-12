package main

import (
	"gopkg.in/urfave/cli.v1"

	// Empty imports to have the init-functions called which should
	// register the protocol
	_ "github.com/JLRgithub/PrivateDCi2b2/protocols"
	_ "github.com/JLRgithub/PrivateDCi2b2/services"
	"gopkg.in/dedis/onet.v1/app"
)

func runServer(ctx *cli.Context) error {
	// first check the options
	config := ctx.String("config")

	app.RunServer(config)

	return nil
}
