package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/paddi-app/paddi/internal/api"
	"github.com/paddi-app/paddi/internal/commands"
	"github.com/paddi-app/paddi/internal/credentials"
)

const (
	exitUserError = 1
	exitAuth      = 2
	exitServer    = 3
)

var version = "0.2.0"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd := commands.Root()
	cmd.Version = version
	if err := cmd.Run(ctx, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "paddi:", err)
		os.Exit(exitCode(err))
	}
}

func exitCode(err error) int {
	var apiErr *api.Error
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.Status == 401 || apiErr.Status == 403:
			return exitAuth
		case apiErr.Status >= 500:
			return exitServer
		default:
			return exitUserError
		}
	}
	if errors.Is(err, credentials.ErrNotLoggedIn) {
		return exitAuth
	}
	var urlErr *url.Error
	var netErr net.Error
	if errors.As(err, &urlErr) || errors.As(err, &netErr) {
		return exitServer
	}
	return exitUserError
}
