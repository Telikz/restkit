package main

import (
	"context"

	"charm.land/fang/v2"
	"github.com/telikz/restkit/cmd/restkit/cli"
)

func main() {
	ctx := context.Background()
	rootCmd := cli.RootCmd{}
	if err := fang.Execute(ctx, rootCmd.Execute(),
		fang.WithColorSchemeFunc(fang.AnsiColorScheme),
	); err != nil {
		return
	}
}
