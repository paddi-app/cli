package commands

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/paddi-app/paddi/internal/api"
	"github.com/paddi-app/paddi/internal/config"
	"github.com/paddi-app/paddi/internal/output"
	"github.com/paddi-app/paddi/internal/prompt"
)

func projectCommand() *cli.Command {
	return &cli.Command{
		Name:  "project",
		Usage: "Manage project context",
		Commands: []*cli.Command{
			{Name: "list", Usage: "List projects in the current workspace", Action: runProjectList},
			{Name: "use", Usage: "Set the current project", ArgsUsage: "[project-id]", Action: runProjectUse},
		},
	}
}

func runProjectList(ctx context.Context, _ *cli.Command) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	projects, raw, err := client.ListProjects(ctx, cfg.Context.WorkspaceID)
	if err != nil {
		return err
	}
	if opts.JSON {
		return output.JSON(os.Stdout, raw)
	}
	if opts.Quiet {
		for _, p := range projects {
			fmt.Println(p.ID)
		}
		return nil
	}
	rows := make([][]string, 0, len(projects))
	for _, p := range projects {
		rows = append(rows, []string{p.ID, p.Name, truncate(p.Description, 60)})
	}
	return output.Table(os.Stdout, []string{"ID", "NAME", "DESCRIPTION"}, rows)
}

func runProjectUse(ctx context.Context, cmd *cli.Command) error {
	proj, err := chooseProject(ctx, cmd)
	if err != nil || proj == nil {
		return err
	}
	if err := config.Set("context.project_id", proj.ID); err != nil {
		return err
	}
	// Adopt the project's workspace so the two stay in sync and subsequent
	// listings are scoped to it.
	if proj.WorkspaceID != "" {
		if err := config.Set("context.workspace_id", proj.WorkspaceID); err != nil {
			return err
		}
	}
	if !opts.Quiet {
		name := proj.Name
		if name == "" {
			name = proj.ID
		}
		fmt.Printf("Project set to %s\n", name)
	}
	return nil
}

// chooseProject resolves the selected project from an explicit argument or an
// interactive picker. The returned project carries its WorkspaceID so callers
// can adopt it as the current workspace. Returns (nil, nil) when the user
// cancels the picker.
func chooseProject(ctx context.Context, cmd *cli.Command) (*api.Project, error) {
	if cmd.Args().Len() > 1 {
		return nil, errors.New("usage: paddi project use [project-id]")
	}

	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}

	// Explicit id: resolve across all workspaces (empty filter) so its
	// workspace can be adopted even when it differs from the current context.
	if id := cmd.Args().First(); id != "" {
		projects, _, err := client.ListProjects(ctx, "")
		if err != nil {
			return nil, err
		}
		for i := range projects {
			if projects[i].ID == id {
				return &projects[i], nil
			}
		}
		return &api.Project{ID: id}, nil
	}

	// No id: present a picker scoped to the current workspace.
	projects, _, err := client.ListProjects(ctx, cfg.Context.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return nil, errors.New("no projects available in the current workspace")
	}

	items := make([]prompt.Item, 0, len(projects))
	for _, p := range projects {
		items = append(items, prompt.Item{ID: p.ID, Label: p.Name})
	}
	choice, err := prompt.Select("Select a project (↑/↓ or j/k to move, Enter to select, q to cancel):", items)
	if errors.Is(err, prompt.ErrCancelled) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	for i := range projects {
		if projects[i].ID == choice.ID {
			return &projects[i], nil
		}
	}
	return &api.Project{ID: choice.ID, Name: choice.Label}, nil
}
