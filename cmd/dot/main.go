package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/yourusername/dot/internal/dotfiles"
	"github.com/yourusername/dot/internal/linker"
)

// Version information (injected by GoReleaser)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.VersionPrinter = func(_ *cli.Command) {
		fmt.Printf("version=%s commit=%s date=%s\n", version, commit, date)
	}
	app := &cli.Command{
		Name:  "dot",
		Usage: "Manage dotfiles with profiles",
		Commands: []*cli.Command{
			checkCmd(),
			cleanCmd(),
			cloneCmd(),
			linkCmd(),
			listCmd(),
			rootCmd(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func checkCmd() *cli.Command {
	return &cli.Command{
		Name:  "check",
		Usage: "Verify that symbolic links defined in the specified profile(s) exist and point to the correct source files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "profile",
				Usage: "Comma-separated list of profiles to check (default: general)",
				Value: "general",
			},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			profiles := linker.ParseProfiles(c.String("profile"))
			return linker.Check(profiles)
		},
	}
}

func cleanCmd() *cli.Command {
	return &cli.Command{
		Name:  "clean",
		Usage: "Remove all registered symbolic links from the home directory as defined in the specified profile(s)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "profile",
				Usage: "Comma-separated list of profiles to clean (default: general)",
				Value: "general",
			},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			profiles := linker.ParseProfiles(c.String("profile"))
			return linker.Clean(profiles)
		},
	}
}

func cloneCmd() *cli.Command {
	return &cli.Command{
		Name:      "clone",
		Usage:     "Clone a dotfiles repository from a remote URL to ~/.dotfiles",
		ArgsUsage: "<repository-url>",
		Action: func(_ context.Context, c *cli.Command) error {
			if c.Args().Len() != 1 {
				return fmt.Errorf("exactly one argument (repository URL) is required")
			}
			return dotfiles.Clone(c.Args().First())
		},
	}
}

func linkCmd() *cli.Command {
	return &cli.Command{
		Name:  "link",
		Usage: "Create symbolic links in the home directory based on the .mappings file for the specified profile(s)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "profile",
				Usage: "Comma-separated list of profiles to link (default: general)",
				Value: "general",
			},
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Usage:   "Simulate link creation without performing I/O operations",
			},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			profiles := linker.ParseProfiles(c.String("profile"))
			dryRun := c.Bool("dry-run")
			return linker.Link(profiles, dryRun)
		},
	}
}

func listCmd() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Show all symbolic links that are currently set based on the specified profile(s)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "profile",
				Usage: "Comma-separated list of profiles to list (default: general)",
				Value: "general",
			},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			profiles := linker.ParseProfiles(c.String("profile"))
			return linker.List(profiles)
		},
	}
}

func rootCmd() *cli.Command {
	return &cli.Command{
		Name:  "root",
		Usage: "Print the dotfiles repository path and exit",
		Action: func(_ context.Context, _ *cli.Command) error {
			return dotfiles.PrintRoot()
		},
	}
}
