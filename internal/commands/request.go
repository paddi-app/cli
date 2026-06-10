package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/urfave/cli/v3"

	"github.com/paddi-app/paddi/internal/api"
	"github.com/paddi-app/paddi/internal/output"
)

func requestCommand() *cli.Command {
	return &cli.Command{
		Name:  "request",
		Usage: "Work with feedback requests",
		Commands: []*cli.Command{
			{Name: "list", Usage: "List requests in the current project, sorted by RIGE score", Action: runRequestList},
			{Name: "view", Usage: "Show a request's analysis, score and solution paths", ArgsUsage: "<request-id>", Action: runRequestView},
			{
				Name:      "regenerate",
				Usage:     "Regenerate a request's solution paths",
				ArgsUsage: "<request-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "expectation", Aliases: []string{"e"}, Usage: "expectation guiding the regeneration"},
				},
				Action: runRequestRegenerate,
			},
			{
				Name:      "draft",
				Usage:     "Answer solution paths and trigger spec generation",
				ArgsUsage: "<request-id>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Required: true, Usage: "answers JSON file (use '-' for stdin)"},
				},
				Action: runRequestDraft,
			},
		},
	}
}

func runRequestList(ctx context.Context, _ *cli.Command) error {
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
	requests, raw, err := client.ListRequests(ctx, projectID)
	if err != nil {
		return err
	}
	if opts.JSON {
		return output.JSON(os.Stdout, raw)
	}
	sort.SliceStable(requests, func(i, j int) bool { return requests[i].Score > requests[j].Score })
	if opts.Quiet {
		for _, r := range requests {
			fmt.Println(r.ID)
		}
		return nil
	}
	rows := make([][]string, 0, len(requests))
	for _, r := range requests {
		rows = append(rows, []string{r.ID, truncate(r.Name, 50), r.Type, r.Status, fmt.Sprintf("%.1f", r.Score)})
	}
	return output.Table(os.Stdout, []string{"ID", "NAME", "TYPE", "STATUS", "SCORE"}, rows)
}

func runRequestView(ctx context.Context, cmd *cli.Command) error {
	id, err := singleArg(cmd, "paddi request view <request-id>")
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
	req, raw, err := client.GetRequest(ctx, id)
	if err != nil {
		return err
	}
	if opts.JSON {
		return output.JSON(os.Stdout, raw)
	}
	printRequest(req)
	return nil
}

func printRequest(r *api.Request) {
	fmt.Printf("%s (%s)\n", r.Name, r.ID)
	fmt.Printf("Type: %s  Status: %s  Score: %.1f\n", r.Type, r.Status, r.Score)
	fmt.Printf("RIGE: reach %.2g x impact %.2g x goal %.2g / effort %.2g\n", r.Reach, r.Impact, r.GoalAlignment, r.Effort)
	if r.Description != "" {
		fmt.Printf("\nDescription:\n%s\n", r.Description)
	}
	if r.Analysis != "" {
		fmt.Printf("\nAnalysis:\n%s\n", r.Analysis)
	}
	if len(r.Captures) > 0 {
		fmt.Printf("\nCaptures: %d\n", len(r.Captures))
	}
	if len(r.SolutionPaths) > 0 {
		fmt.Println("\nSolution paths:")
		for i, p := range r.SolutionPaths {
			multiple := ""
			if p.Multiple {
				multiple = " [multiple]"
			}
			fmt.Printf("%d. %s%s (%s)\n", i+1, p.Question, multiple, p.ID)
			if p.Context != "" {
				fmt.Printf("   Context: %s\n", p.Context)
			}
			if p.Impact != "" {
				fmt.Printf("   Impact: %s\n", p.Impact)
			}
			for _, o := range p.Options {
				fmt.Printf("   - %s — %s\n", o.Label, o.Impact)
			}
			for _, sel := range p.Selections {
				custom := ""
				if sel.Custom {
					custom = " (custom)"
				}
				fmt.Printf("   > selected: %s%s\n", sel.Label, custom)
			}
		}
	}
	if r.Spec != nil {
		locked := ""
		if r.Spec.Locked {
			locked = ", locked"
		}
		fmt.Printf("\nSpec: %s (%s%s)\n", r.Spec.Title, r.Spec.ID, locked)
	}
}

func runRequestRegenerate(ctx context.Context, cmd *cli.Command) error {
	id, err := singleArg(cmd, "paddi request regenerate <request-id> [-e <expectation>]")
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
	req, raw, err := client.RegenerateRequest(ctx, id, cmd.String("expectation"))
	if err != nil {
		return err
	}
	if opts.JSON {
		return output.JSON(os.Stdout, raw)
	}
	if !opts.Quiet {
		fmt.Printf("Regenerating solution paths for %q (status: %s)\n", req.Name, req.Status)
	}
	return nil
}

func runRequestDraft(ctx context.Context, cmd *cli.Command) error {
	id, err := singleArg(cmd, "paddi request draft <request-id> -f <answers.json>")
	if err != nil {
		return err
	}
	answers, err := readAnswers(cmd.String("file"))
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
	req, raw, err := client.DraftRequest(ctx, id, answers)
	if err != nil {
		return err
	}
	if opts.JSON {
		return output.JSON(os.Stdout, raw)
	}
	if !opts.Quiet {
		fmt.Printf("Drafting spec for %q (status: %s)\n", req.Name, req.Status)
	}
	return nil
}

// readAnswers accepts either a bare answers array or a {"answers": [...]} object.
func readAnswers(path string) ([]api.Answer, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}

	data = bytes.TrimSpace(data)
	var answers []api.Answer
	if len(data) > 0 && data[0] == '[' {
		err = json.Unmarshal(data, &answers)
	} else {
		var body struct {
			Answers []api.Answer `json:"answers"`
		}
		err = json.Unmarshal(data, &body)
		answers = body.Answers
	}
	if err != nil {
		return nil, fmt.Errorf("invalid answers JSON: %w", err)
	}
	if len(answers) == 0 {
		return nil, errors.New("answers must contain at least one entry")
	}
	return answers, nil
}
