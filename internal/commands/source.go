package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/paddi-app/paddi/internal/output"
)

func sourceCommand() *cli.Command {
	return &cli.Command{
		Name:  "source",
		Usage: "Manage data sources",
		Commands: []*cli.Command{
			{Name: "list", Usage: "List data sources in the current project", Action: runSourceList},
			{Name: "index", Usage: "Trigger re-indexing of a source", ArgsUsage: "<source-id>", Action: runSourceIndex},
		},
	}
}

func runSourceList(ctx context.Context, _ *cli.Command) error {
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
	sources, raw, err := client.ListSources(ctx, projectID)
	if err != nil {
		return err
	}
	if opts.JSON {
		return output.JSON(os.Stdout, raw)
	}
	if opts.Quiet {
		for _, s := range sources {
			fmt.Println(s.ID)
		}
		return nil
	}
	rows := make([][]string, 0, len(sources))
	for _, s := range sources {
		name := s.DisplayName
		if name == "" {
			name = s.Subject
		}
		status := s.LookupStatus
		if status == "" {
			status = "-"
		}
		lastIndexed := "never"
		if !s.LastIndexedAt.IsZero() {
			lastIndexed = s.LastIndexedAt.Local().Format("2006-01-02 15:04")
		}
		rows = append(rows, []string{s.ID, s.Provider, s.Type, truncate(name, 40), status, lastIndexed})
	}
	return output.Table(os.Stdout, []string{"ID", "PROVIDER", "TYPE", "NAME", "STATUS", "LAST INDEXED"}, rows)
}

func runSourceIndex(ctx context.Context, cmd *cli.Command) error {
	id, err := singleArg(cmd, "paddi source index <source-id>")
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
	if err := client.IndexSource(ctx, id); err != nil {
		return err
	}
	if !opts.Quiet {
		fmt.Printf("Indexing triggered for source %s\n", id)
	}
	return nil
}
