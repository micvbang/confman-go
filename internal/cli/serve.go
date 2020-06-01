package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/micvbang/confman-go/internal/httpapi"
	"github.com/micvbang/confman-go/internal/httpapi/httphelpers"
	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ServeCommandInput struct {
	Addr string
	Port string
}

func ConfigureServeCommand(ctx context.Context, app *kingpin.Application, log logger.Logger) {
	input := ServeCommandInput{}

	cmd := app.Command("serve", "Serves a web interface to visualize data")
	cmd.Arg("addr", "Address to listen on").
		Default("127.0.0.1").
		StringVar(&input.Addr)

	cmd.Arg("port", "Port to listen on").
		Default("8000").
		StringVar(&input.Port)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(ServeCommand(ctx, input, os.Stdout, log, GlobalFlags.Storage), "serve")
		return nil
	})
}

func ServeCommand(ctx context.Context, input ServeCommandInput, w io.Writer, log logger.Logger, storage storage.Storage) error {
	httpDependencies := httpapi.Dependencies{
		Storage: storage,
	}
	flags := httpapi.Flags{
		ListenAddr: input.Addr,
		ListenPort: input.Port,
	}

	fmt.Fprintf(w, "Listening on http://%v:%v\n", input.Addr, input.Port)
	return httpapi.ListenAndServe(flags, log, func(r httphelpers.Router) httphelpers.Router {
		return httpapi.AddRoutes(r, httpDependencies)
	})
}
