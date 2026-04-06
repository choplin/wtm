package main

import (
	"fmt"
	"os"

	"github.com/choplin/wtm/internal/wtm"
	"github.com/spf13/cobra"
)

var version = "dev"

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
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newAddCmd(),
		newListCmd(),
		newShowCmd(),
		newRemoveCmd(),
		newNotesCmd(),
		newConfigCmd(),
		newVersionCmd(),
		newCompletionCmd(),
	)

	return cmd
}

func newAddCmd() *cobra.Command {
	var branch string
	var checkout string
	var base string
	var message string

	cmd := &cobra.Command{
		Use:   "add [<name>]",
		Short: "Create a new worktree",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			if err := wtm.AddWorktree(name, branch, checkout, base, message); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Create new branch with specified name")
	cmd.Flags().StringVarP(&checkout, "checkout", "B", "", "Use existing branch")
	cmd.Flags().StringVar(&base, "base", "", "Base branch for new branch")
	cmd.Flags().StringVarP(&message, "message", "m", "", "Attach a note to the worktree")

	return cmd
}

func newListCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all worktrees",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := wtm.ListWorktrees(format); err != nil {
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
		Use:               "show <name>",
		Short:             "Show worktree details",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeWorktreeNames,
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			if err := wtm.ShowWorktree(name, format, field); err != nil {
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
		Use:               "remove <name>",
		Short:             "Remove a worktree",
		Aliases:           []string{"rm"},
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeWorktreeNames,
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]

			if deleteBranch && deleteBranchForce {
				return fmt.Errorf("cannot combine --delete-branch and --delete-branch-force")
			}

			opts := wtm.RemoveOptions{Force: force}
			switch {
			case deleteBranch:
				opts.BranchDelete = wtm.BranchDeleteSafe
			case deleteBranchForce:
				opts.BranchDelete = wtm.BranchDeleteForce
			}

			if err := wtm.RemoveWorktree(name, opts); err != nil {
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

func newNotesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notes",
		Short: "Manage worktree notes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newNotesAddCmd(),
		newNotesShowCmd(),
		newNotesEditCmd(),
		newNotesRemoveCmd(),
	)

	return cmd
}

func newNotesAddCmd() *cobra.Command {
	var message string
	var force bool

	cmd := &cobra.Command{
		Use:               "add <worktree>",
		Short:             "Add a note to a worktree",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeWorktreeNames,
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			if message == "" {
				return wtm.EditNote(name)
			}
			return wtm.AddNote(name, message, force)
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Note message")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing note")

	return cmd
}

func newNotesShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "show <worktree>",
		Short:             "Show a worktree note",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeWorktreeNames,
		RunE: func(_ *cobra.Command, args []string) error {
			note, err := wtm.GetNote(args[0])
			if err != nil {
				return err
			}
			if note == "" {
				return fmt.Errorf("no note found for worktree '%s'", args[0])
			}
			fmt.Println(note)
			return nil
		},
	}
}

func newNotesEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "edit <worktree>",
		Short:             "Edit a worktree note in $EDITOR",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeWorktreeNames,
		RunE: func(_ *cobra.Command, args []string) error {
			return wtm.EditNote(args[0])
		},
	}
}

func newNotesRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "remove <worktree>",
		Short:             "Remove a worktree note",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeWorktreeNames,
		RunE: func(_ *cobra.Command, args []string) error {
			return wtm.RemoveNote(args[0])
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("wtm version %s\n", version)
		},
	}
}

func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate shell completion script",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage project configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newConfigEditCmd())

	return cmd
}

func newConfigEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit project configuration in your editor",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return wtm.EditConfig()
		},
	}
}

func completeWorktreeNames(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	worktrees, err := wtm.GetWorktrees()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var names []string
	for _, wt := range worktrees {
		names = append(names, fmt.Sprintf("%s\t%s", wt.Name, wt.Branch))
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}
