package commands

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/paddi-app/cli/internal/api"
	"github.com/paddi-app/cli/internal/output"
)

func specCommand() *cli.Command {
	return &cli.Command{
		Name:  "spec",
		Usage: "Work with specs",
		Commands: []*cli.Command{
			{Name: "list", Usage: "List specs in the current project", Action: runSpecList},
			{Name: "view", Usage: "Print a spec's markdown content", ArgsUsage: "<spec-id>", Action: runSpecView},
			{
				Name:      "download",
				Usage:     "Write a spec's markdown content to a local file",
				ArgsUsage: "<spec-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "output file path (default: <title>.md)"},
				},
				Action: runSpecDownload,
			},
			{Name: "lock", Usage: "Lock a spec to prevent further edits", ArgsUsage: "<spec-id>", Action: runSpecLock},
		},
	}
}

func runSpecList(ctx context.Context, _ *cli.Command) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	projectID, err := requireProject(cfg)
	if err != nil {
		return err
	}
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	specs, raw, err := client.ListSpecs(ctx, projectID)
	if err != nil {
		return err
	}
	if opts.JSON {
		return output.JSON(os.Stdout, raw)
	}
	sort.SliceStable(specs, func(i, j int) bool { return specs[i].CreatedAt.After(specs[j].CreatedAt) })
	if opts.Quiet {
		for _, s := range specs {
			fmt.Println(s.ID)
		}
		return nil
	}
	rows := make([][]string, 0, len(specs))
	for _, s := range specs {
		requestType := ""
		if s.Request != nil {
			requestType = s.Request.Type
		}
		locked := "no"
		if s.Locked {
			locked = "yes"
		}
		rows = append(rows, []string{s.ID, truncate(s.Title, 50), requestType, locked, s.CreatedAt.Local().Format("2006-01-02")})
	}
	return output.Table(os.Stdout, []string{"ID", "TITLE", "REQUEST TYPE", "LOCKED", "CREATED"}, rows)
}

func runSpecView(ctx context.Context, cmd *cli.Command) error {
	id, err := singleArg(cmd, "paddi spec view <spec-id>")
	if err != nil {
		return err
	}
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	spec, raw, err := client.GetSpec(ctx, id)
	if err != nil {
		return err
	}
	if opts.JSON {
		return output.JSON(os.Stdout, raw)
	}
	fmt.Printf("# %s\n", spec.Title)
	fmt.Print(spec.Content)
	if !strings.HasSuffix(spec.Content, "\n") {
		fmt.Println()
	}
	return nil
}

func runSpecDownload(ctx context.Context, cmd *cli.Command) error {
	id, err := singleArg(cmd, "paddi spec download <spec-id> [-o <path>]")
	if err != nil {
		return err
	}
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	spec, _, err := client.GetSpec(ctx, id)
	if err != nil {
		return err
	}
	path := cmd.String("output")
	if path == "" {
		path = specFilename(spec)
	}
	if err := os.WriteFile(path, []byte(spec.Content), 0o644); err != nil {
		return err
	}
	if opts.Quiet {
		fmt.Println(path)
	} else {
		fmt.Printf("Wrote spec %q to %s\n", spec.Title, path)
	}
	return nil
}

func runSpecLock(ctx context.Context, cmd *cli.Command) error {
	id, err := singleArg(cmd, "paddi spec lock <spec-id>")
	if err != nil {
		return err
	}
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	spec, raw, err := client.LockSpec(ctx, id)
	if err != nil {
		return err
	}
	if opts.JSON {
		return output.JSON(os.Stdout, raw)
	}
	if !opts.Quiet {
		fmt.Printf("Spec %q locked.\n", spec.Title)
	}
	return nil
}

func specFilename(s *api.Spec) string {
	title := strings.TrimSpace(s.Title)
	if title == "" {
		return s.ID + ".md"
	}
	repl := strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-", "*", "-", "?", "-",
		"\"", "-", "<", "-", ">", "-", "|", "-",
	)
	return repl.Replace(title) + ".md"
}
