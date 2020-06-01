package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/micvbang/confman-go/internal/cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	run(os.Args[1:], os.Exit)
}

var Version = "dev"

func run(args []string, exit func(int)) {
	app := kingpin.New(
		"confman",
		" A tool for easily managing configurations for services",
	)

	app.ErrorWriter(os.Stderr)
	app.Writer(os.Stdout)
	app.Version(Version)
	app.Terminate(exit)

	ctx := setupSigIntHandler()

	log := cli.ConfigureGlobals(app)
	cli.ConfigureReadCommand(ctx, app, log)
	cli.ConfigureWriteCommand(ctx, app, log)
	cli.ConfigureListCommand(ctx, app, log)
	cli.ConfigureDeleteCommand(ctx, app, log)
	cli.ConfigureExecCommand(ctx, app, log)
	cli.ConfigureDeployCommand(ctx, app, log)
	cli.ConfigureServeCommand(ctx, app, log)

	kingpin.MustParse(app.Parse(args))
}

func setupSigIntHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func() {
		<-c
		fmt.Fprintf(os.Stdout, "Terminating...")
		cancel()
		os.Exit(1)
	}()

	return ctx
}
