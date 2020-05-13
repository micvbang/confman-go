package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

const (
	formatJSON = "json"
	formatText = "txt"
	formatYAML = "yaml"
)

func addFlagOutputFormat(cmd *kingpin.CmdClause, format *string) {
	cmd.Flag("format", "Format of output").
		Short('f').
		Default(formatText).
		Envar("CONFMAN_DEFAULT_FORMAT").
		EnumVar(format, formatText, formatJSON, formatYAML)
}

type outputFormatter func(w io.Writer, v interface{}) error

var outputFormatters map[string]outputFormatter = map[string]outputFormatter{
	formatJSON: outputJSON,
	formatYAML: outputYAML,
}

func outputFormat(format string, w io.Writer, v interface{}) error {
	f, exists := outputFormatters[format]
	if !exists {
		return fmt.Errorf("formatter \"%s\" does not exist", format)
	}

	return f(w, v)
}

func outputJSON(w io.Writer, v interface{}) error {
	bs, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	w.Write(bs)
	return nil
}

func outputYAML(w io.Writer, v interface{}) error {
	bs, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	w.Write(bs)
	return nil
}
