package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/acronis/go-stacktrace"
	"github.com/acronis/go-stacktrace/slogex"
	"github.com/spf13/cobra"
)

type CommandError struct {
	Inner error
	Msg   string
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Inner)
}

func (e *CommandError) Unwrap() error {
	return e.Inner
}

func NewCommandError(err error, msg string) error {
	if err != nil {
		return &CommandError{Inner: err, Msg: msg}
	}
	return nil
}

type Command interface {
	Execute(ctx context.Context) error
}

func InitLoggingAndRun(ctx context.Context, verbosity int, cmd Command) error {
	lvl := slog.LevelInfo
	if verbosity > 0 {
		lvl = slog.LevelDebug
	}
	InitLogging(lvl)
	return NewCommandError(cmd.Execute(ctx), "command error")
}

func main() {
	os.Exit(mainFn())
}

func mainFn() int {
	var ensureDuplicates bool
	verbosity := 0
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	cmdValidate := func() *cobra.Command {
		opts := ValidateOptions{
			EnsureDuplicates: ensureDuplicates,
		}
		cmd := &cobra.Command{
			Use:   "validate",
			Short: "validate raml files",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(_ *cobra.Command, args []string) error {
				return InitLoggingAndRun(ctx, verbosity, NewValidateCmd(opts, args))
			},
		}

		return cmd
	}()

	rootCmd := func() *cobra.Command {
		cmd := &cobra.Command{
			Use:           "raml",
			Short:         "raml is a RAML 1.0 tool",
			SilenceUsage:  true,
			SilenceErrors: true,
			CompletionOptions: cobra.CompletionOptions{
				DisableDefaultCmd: true,
			},
		}

		cmd.PersistentFlags().CountVarP(&verbosity, "verbosity", "v", "increase verbosity level: -v for debug")
		cmd.Flags().BoolVarP(&ensureDuplicates, "ensure-duplicates", "d", false,
			"ensure that there are no duplicates in tracebacks")

		cmd.AddCommand(
			cmdValidate,
		)
		return cmd
	}()

	if err := rootCmd.Execute(); err != nil {
		var cmdErr *CommandError
		if errors.As(err, &cmdErr) && cmdErr.Inner != nil {
			stOpts := []stacktrace.TracesOpt{}
			if ensureDuplicates {
				stOpts = append(stOpts, stacktrace.WithEnsureDuplicates())
			}
			slog.Error("Command failed", slogex.ErrToSlogAttr(cmdErr.Inner, stOpts...))
		} else {
			_ = rootCmd.Usage()
		}
		return 1
	}

	return 0
}
