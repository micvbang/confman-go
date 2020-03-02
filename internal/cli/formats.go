package cli

import (
	"encoding/json"
	"io"

	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	formatJSON = "json"
	formatText = "txt"
)

func addFlagOutputFormat(cmd *kingpin.CmdClause, format *string) {
	cmd.Flag("format", "Format of output").
		Short('f').
		Default(formatText).
		Envar("CONFMAN_DEFAULT_FORMAT").
		EnumVar(format, formatText, formatJSON)
}

func outputJSON(w io.Writer, serviceName string, v interface{}) error {
	output := map[string]interface{}{
		serviceName: v,
	}
	bs, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	w.Write(bs)
	return nil
}
