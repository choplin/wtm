package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.1"

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "wtm",
		Short:         "Worktree Manager",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newAddCmd(),
		newListCmd(),
		newShowCmd(),
		newRemoveCmd(),
		newVersionCmd(),
		newMCPCmd(),
	)

	return cmd
}

func newAddCmd() *cobra.Command {
	var branch string
	var checkout string
	var base string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := AddWorktree(name, branch, checkout, base); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Create new branch with specified name")
	cmd.Flags().StringVarP(&checkout, "checkout", "B", "", "Use existing branch")
	cmd.Flags().StringVar(&base, "base", "", "Base branch for new branch")

	return cmd
}

func newListCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all worktrees",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ListWorktrees(format); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "Output format: table, plain, json")

	return cmd
}

func newShowCmd() *cobra.Command {
	var format string
	var field string

	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show worktree details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := ShowWorktree(name, format, field); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "pretty", "Output format: pretty, json")
	cmd.Flags().StringVarP(&field, "field", "f", "", "Output specific field only")

	return cmd
}

func newRemoveCmd() *cobra.Command {
	var force bool
	var deleteBranch bool
	var deleteBranchForce bool

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if deleteBranch && deleteBranchForce {
				return fmt.Errorf("cannot combine --delete-branch and --delete-branch-force")
			}

			opts := RemoveOptions{Force: force}
			switch {
			case deleteBranch:
				opts.BranchDelete = BranchDeleteSafe
			case deleteBranchForce:
				opts.BranchDelete = BranchDeleteForce
			}

			if err := RemoveWorktree(name, opts); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	cmd.Flags().BoolVarP(&deleteBranch, "delete-branch", "d", false, "Delete associated branch (git branch -d)")
	cmd.Flags().BoolVarP(&deleteBranchForce, "delete-branch-force", "D", false, "Force delete associated branch (git branch -D)")
	cmd.MarkFlagsMutuallyExclusive("delete-branch", "delete-branch-force")

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("wtm version %s\n", version)
		},
	}
}

func newMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			if err := StartMCPServer(ctx); err != nil {
				return err
			}
			return nil
		},
	}
}
