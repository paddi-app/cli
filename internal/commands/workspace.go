package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/paddi-app/cli/internal/api"
	"github.com/paddi-app/cli/internal/config"
	"github.com/paddi-app/cli/internal/output"
	"github.com/paddi-app/cli/internal/prompt"
)

func workspaceCommand() *cli.Command {
	return &cli.Command{
		Name:  "workspace",
		Usage: "Manage workspace context",
		Commands: []*cli.Command{
			{Name: "list", Usage: "List my workspaces", Action: runWorkspaceList},
			{Name: "use", Usage: "Set the current workspace", ArgsUsage: "[workspace-id]", Action: runWorkspaceUse},
		},
	}
}

func runWorkspaceList(ctx context.Context, _ *cli.Command) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	workspaces, raw, err := client.ListWorkspaces(ctx)
	if err != nil {
		return err
	}
	if opts.JSON {
		return output.JSON(os.Stdout, raw)
	}
	if opts.Quiet {
		for _, w := range workspaces {
			fmt.Println(w.ID)
		}
		return nil
	}
	rows := make([][]string, 0, len(workspaces))
	for _, w := range workspaces {
		rows = append(rows, []string{w.ID, w.Name, roleName(w.Member), strconv.Itoa(len(w.Projects))})
	}
	return output.Table(os.Stdout, []string{"ID", "NAME", "ROLE", "PROJECTS"}, rows)
}

func runWorkspaceUse(ctx context.Context, cmd *cli.Command) error {
	id, name, err := chooseWorkspace(ctx, cmd)
	if err != nil || id == "" {
		return err
	}
	if err := config.Set("context.workspace_id", id); err != nil {
		return err
	}
	if !opts.Quiet {
		if name == "" {
			name = id
		}
		fmt.Printf("Workspace set to %s\n", name)
	}
	return nil
}

// chooseWorkspace resolves the workspace id from an explicit argument, or by
// presenting an interactive picker when none is given. The second return value
// is the workspace name (empty when an id is passed directly). Both are empty
// (with a nil error) when the user cancels the picker.
func chooseWorkspace(ctx context.Context, cmd *cli.Command) (id, name string, err error) {
	switch cmd.Args().Len() {
	case 1:
		return cmd.Args().First(), "", nil
	case 0:
	default:
		return "", "", errors.New("usage: paddi workspace use [workspace-id]")
	}

	cfg, err := loadConfig()
	if err != nil {
		return "", "", err
	}
	client, err := newClient(cfg)
	if err != nil {
		return "", "", err
	}
	workspaces, _, err := client.ListWorkspaces(ctx)
	if err != nil {
		return "", "", err
	}
	if len(workspaces) == 0 {
		return "", "", errors.New("no workspaces available")
	}

	items := make([]prompt.Item, 0, len(workspaces))
	for _, w := range workspaces {
		items = append(items, prompt.Item{ID: w.ID, Label: w.Name})
	}
	choice, err := prompt.Select("Select a workspace (↑/↓ or j/k to move, Enter to select, q to cancel):", items)
	if errors.Is(err, prompt.ErrCancelled) {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}
	return choice.ID, choice.Label, nil
}

func roleName(m *api.WorkspaceMember) string {
	if m == nil {
		return ""
	}
	switch m.Role {
	case 1:
		return "Member"
	case 2:
		return "Admin"
	case 3:
		return "Owner"
	default:
		return strconv.Itoa(m.Role)
	}
}
