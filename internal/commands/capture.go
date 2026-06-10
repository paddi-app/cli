package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/paddi-app/paddi/internal/api"
	"github.com/paddi-app/paddi/internal/output"
)

func captureCommand() *cli.Command {
	return &cli.Command{
		Name:  "capture",
		Usage: "Feed raw feedback into Paddi",
		Commands: []*cli.Command{
			{
				Name:      "create",
				Usage:     "Create a capture from a message, file, or stdin",
				ArgsUsage: "[-]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "message", Aliases: []string{"m"}, Usage: "feedback text"},
					&cli.StringFlag{Name: "file", Aliases: []string{"f"}, Usage: "read feedback from a file"},
					&cli.StringFlag{Name: "origin", Usage: "origin id"},
					&cli.StringSliceFlag{Name: "tag", Usage: "tag name (repeatable)"},
				},
				Action: runCaptureCreate,
			},
		},
	}
}

func runCaptureCreate(ctx context.Context, cmd *cli.Command) error {
	description, err := captureDescription(cmd)
	if err != nil {
		return err
	}
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
	capture, raw, err := client.CreateCapture(ctx, api.CaptureInput{
		ProjectID:   projectID,
		Description: description,
		OriginID:    cmd.String("origin"),
		Tags:        cmd.StringSlice("tag"),
	})
	if err != nil {
		return err
	}
	if opts.JSON {
		return output.JSON(os.Stdout, raw)
	}
	if opts.Quiet {
		fmt.Println(capture.ID)
	} else {
		fmt.Printf("Capture %s created.\n", capture.ID)
	}
	return nil
}

func captureDescription(cmd *cli.Command) (string, error) {
	switch {
	case cmd.String("message") != "":
		return cmd.String("message"), nil
	case cmd.String("file") != "" && cmd.String("file") != "-":
		data, err := os.ReadFile(cmd.String("file"))
		if err != nil {
			return "", err
		}
		return validDescription(string(data))
	case cmd.String("file") == "-" || cmd.Args().First() == "-":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return validDescription(string(data))
	default:
		return "", errors.New("provide feedback via -m <message>, -f <file>, or '-' for stdin")
	}
}

func validDescription(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errors.New("feedback is empty")
	}
	return s, nil
}
