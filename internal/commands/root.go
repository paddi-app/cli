package commands

import (
	"context"
	"errors"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/paddi-app/cli/internal/api"
	"github.com/paddi-app/cli/internal/config"
	"github.com/paddi-app/cli/internal/credentials"
)

type globalOptions struct {
	JSON    bool
	Quiet   bool
	Project string
	APIBase string
}

var opts globalOptions

// Root builds the paddi root command.
func Root() *cli.Command {
	opts = globalOptions{}
	cli.VersionFlag = &cli.BoolFlag{
		Name:        "version",
		Aliases:     []string{"V"},
		Usage:       "print the version",
		HideDefault: true,
		Local:       true,
	}
	return &cli.Command{
		Name:  "paddi",
		Usage: "Paddi from the command line",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json", Usage: "output raw API JSON", Destination: &opts.JSON},
			&cli.BoolFlag{Name: "quiet", Aliases: []string{"q"}, Usage: "only output essential results", Destination: &opts.Quiet},
			&cli.StringFlag{Name: "project", Usage: "override the current project context", Destination: &opts.Project},
			&cli.StringFlag{Name: "api-base", Usage: "override the API base URL", Destination: &opts.APIBase},
		},
		Commands: []*cli.Command{
			authCommand(),
			workspaceCommand(),
			projectCommand(),
			specCommand(),
			requestCommand(),
			captureCommand(),
			sourceCommand(),
		},
	}
}

// loadConfig returns the effective config with flag overrides applied
// (flag > env > file > default).
func loadConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if opts.APIBase != "" {
		cfg.APIBase = opts.APIBase
	}
	if opts.Project != "" {
		cfg.Context.ProjectID = opts.Project
	}
	return cfg, nil
}

// newClient builds an authenticated API client with automatic token refresh.
func newClient(cfg *config.Config) (*api.Client, error) {
	token, err := credentials.AccessToken()
	if err != nil {
		return nil, err
	}
	c := &api.Client{BaseURL: cfg.APIBase, Token: token}
	if !credentials.FromEnv() {
		c.Refresh = func(ctx context.Context) (string, error) {
			rt, err := credentials.RefreshToken()
			if err != nil {
				return "", err
			}
			tokens, err := c.RefreshToken(ctx, rt)
			if err != nil {
				return "", err
			}
			if err := credentials.Store(tokens.AccessToken, tokens.RefreshToken); err != nil {
				return "", err
			}
			return tokens.AccessToken, nil
		}
	}
	return c, nil
}

func requireProject(cfg *config.Config) (string, error) {
	if cfg.Context.ProjectID == "" {
		return "", errors.New("no project selected: run `paddi project use <id>` or pass --project")
	}
	return cfg.Context.ProjectID, nil
}

func singleArg(cmd *cli.Command, usage string) (string, error) {
	if cmd.Args().Len() != 1 {
		return "", errors.New("usage: " + usage)
	}
	return cmd.Args().First(), nil
}

func orNone(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}

func truncate(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
