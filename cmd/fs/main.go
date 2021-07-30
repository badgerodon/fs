package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/badgerodon/fs"
	_ "github.com/badgerodon/fs/providers/http"
	_ "github.com/badgerodon/fs/providers/os"
	_ "github.com/badgerodon/fs/providers/yandex"
	"github.com/mattn/go-isatty"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
)

var isInteractive = isatty.IsTerminal(os.Stdout.Fd())

func main() {
	if isInteractive {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	cpCmd := &ffcli.Command{
		Name:       "cp",
		ShortUsage: "fs cp dst src",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("src is required")
			}

			if len(args) < 2 {
				return fmt.Errorf("dst is required")
			}

			return cp(ctx, args[0], args[1])
		},
	}
	rootCmd := &ffcli.Command{
		ShortUsage:  "fs <cmd>",
		Subcommands: []*ffcli.Command{cpCmd},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}

	err := rootCmd.ParseAndRun(context.Background(), os.Args[1:])
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		log.Fatal().Err(err).Send()
	}
}

func cp(ctx context.Context, rawDstURL, rawSrcURL string) error {
	dstURL, err := url.Parse(rawDstURL)
	if err != nil {
		return fmt.Errorf("invalid destination url: %w", err)
	}
	srcURL, err := url.Parse(rawSrcURL)
	if err != nil {
		return fmt.Errorf("invalid source url: %w", err)
	}

	srcInfo, err := fs.Stat(ctx, srcURL)
	if err != nil {
		return fmt.Errorf("error getting source info: %w", err)
	}

	if isInteractive {
		bar := progressbar.DefaultBytes(srcInfo.Size())
		ctx = fs.WithProgressFunc(context.Background(), func(add int) {
			_ = bar.Add(add)
		})
		defer func() { _ = bar.Close() }()
	}

	err = fs.Copy(ctx, dstURL, srcURL)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	return nil
}
