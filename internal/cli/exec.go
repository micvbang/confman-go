package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"gitlab.com/micvbang/confman-go/pkg/confman"
	"gitlab.com/micvbang/confman-go/pkg/storage"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ExecCommandInput struct {
	ServiceNames string
	Command      string
	Args         []string
}

func ConfigureExecCommand(ctx context.Context, app *kingpin.Application, log confman.Logger, storage storage.Storage) {
	input := ExecCommandInput{}

	cmd := app.Command("exec", "Populates the environment with secrets from the given configurations")
	cmd.Arg("service", "Name of the service(s)").
		Required().
		StringVar(&input.ServiceNames)

	cmd.Arg("cmd", "Command to execute, defaults to $SHELL").
		Default(os.Getenv("SHELL")).
		StringVar(&input.Command)

	cmd.Arg("args", "Command arguments").
		StringsVar(&input.Args)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(ExecCommand(ctx, app, input, log, storage), "exec")
		return nil
	})
}

func ExecCommand(ctx context.Context, app *kingpin.Application, input ExecCommandInput, log confman.Logger, storage storage.Storage) error {
	config := make(map[string]string)

	for _, serviceName := range confman.ParseServicePaths(input.ServiceNames) {
		cm := confman.New(log, storage, serviceName)
		curConfig, err := cm.ReadAll(ctx)
		if err != nil {
			return nil
		}

		for key, value := range curConfig {
			config[key] = value
		}
	}

	env := environ(os.Environ())
	for key, value := range config {
		overwritten := env.Set(key, value)
		if overwritten {
			fmt.Printf("warning: overwriting var %s\n", key)
		}
	}

	return execSyscall(input.Command, input.Args, env)
}

// environ is a slice of strings representing the environment, in the form "key=value".
type environ []string

// Unset an environment variable by key
func (e *environ) Unset(key string) {
	for i := range *e {
		if strings.HasPrefix((*e)[i], key+"=") {
			(*e)[i] = (*e)[len(*e)-1]
			*e = (*e)[:len(*e)-1]
			break
		}
	}
}

func (e *environ) Contains(key string) bool {
	for i := range *e {
		if strings.HasPrefix((*e)[i], key+"=") {
			return true
		}
	}
	return false
}

// Set adds an environment variable, replacing any existing ones of the same key
func (e *environ) Set(key, val string) (overwritten bool) {
	exists := e.Contains(key)
	e.Unset(key)
	*e = append(*e, key+"="+val)
	return exists
}

func supportsExecSyscall() bool {
	return runtime.GOOS == "linux" || runtime.GOOS == "darwin" || runtime.GOOS == "freebsd"
}

func execCmd(command string, args []string, env []string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Failed to start command: %v", err)
	}

	go func() {
		for {
			sig := <-sigChan
			cmd.Process.Signal(sig)
		}
	}()

	if err := cmd.Wait(); err != nil {
		cmd.Process.Signal(os.Kill)
		return fmt.Errorf("Failed to wait for command termination: %v", err)
	}

	waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
	os.Exit(waitStatus.ExitStatus())
	return nil
}

func execSyscall(command string, args []string, env []string) error {
	if !supportsExecSyscall() {
		return execCmd(command, args, env)
	}

	argv0, err := exec.LookPath(command)
	if err != nil {
		return err
	}

	argv := make([]string, 0, 1+len(args))
	argv = append(argv, command)
	argv = append(argv, args...)

	return syscall.Exec(argv0, argv, env)
}
