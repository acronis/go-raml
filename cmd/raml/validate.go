package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/acronis/go-raml/v2"
	"github.com/acronis/go-stacktrace"
	"github.com/acronis/go-stacktrace/slogex"
)

type ValidateOptions struct {
	EnsureDuplicates bool
}

type ValidateCommand struct {
	Opts ValidateOptions
	Args []string
}

func NewValidateCmd(opts ValidateOptions, args []string) *ValidateCommand {
	return &ValidateCommand{
		Opts: opts,
		Args: args,
	}
}

func (v ValidateCommand) Execute(ctx context.Context) error {
	var err error
	var stOpts []stacktrace.TracesOpt
	if v.Opts.EnsureDuplicates {
		stOpts = append(stOpts, stacktrace.WithEnsureDuplicates())
	}
	for _, arg := range v.Args {
		slog.Info("Validating RAML...", slog.String("path", arg))
		_, err = raml.ParseFromPathCtx(ctx, arg, raml.OptWithUnwrap(), raml.OptWithValidate())
		if err != nil {
			slog.Error("RAML is invalid", slogex.ErrToSlogAttr(err, stOpts...))
		} else {
			slog.Info("RAML is valid", slog.String("path", arg))
		}
	}
	if err != nil {
		return fmt.Errorf("errors have been found in the RAML files")
	}
	return nil
}
