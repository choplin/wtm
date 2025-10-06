package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

const version = "0.1.1"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		handleAdd(os.Args[2:])
	case "list":
		handleList(os.Args[2:])
	case "show":
		handleShow(os.Args[2:])
	case "remove":
		handleRemove(os.Args[2:])
	case "version":
		fmt.Printf("wtm version %s\n", version)
	case "mcp":
		handleMCP()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `wtm - Worktree Manager

Usage:
  wtm add <name> [options]     Create a new worktree
  wtm list [options]           List all worktrees
  wtm show <name> [options]    Show worktree details
  wtm remove <name> [options]  Remove a worktree
  wtm version                  Show version information
  wtm mcp                      Start MCP server

Options:
  wtm add:
    -b, --branch <name>    Create new branch with specified name
    -B, --checkout <name>  Use existing branch
    --base <branch>        Base branch for new branch (default: current HEAD)

  wtm list:
    --format <type>        Output format: table (default), plain, json

  wtm show:
    --format <type>        Output format: pretty (default), json
    -f, --field <name>     Output specific field only

  wtm remove:
    -f, --force            Skip confirmation

For more information, visit: https://github.com/akitenkgen/worktree-manager
`)
}

func handleAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	branch := fs.String("b", "", "Create new branch with specified name")
	fs.StringVar(branch, "branch", "", "Create new branch with specified name")
	checkout := fs.String("B", "", "Use existing branch")
	fs.StringVar(checkout, "checkout", "", "Use existing branch")
	base := fs.String("base", "", "Base branch for new branch")

	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: worktree name required")
		os.Exit(1)
	}

	name := fs.Arg(0)

	if err := AddWorktree(name, *branch, *checkout, *base); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	format := fs.String("format", "table", "Output format: table, plain, json")

	fs.Parse(args)

	if err := ListWorktrees(*format); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleShow(args []string) {
	fs := flag.NewFlagSet("show", flag.ExitOnError)
	format := fs.String("format", "pretty", "Output format: pretty, json")
	field := fs.String("f", "", "Output specific field only")
	fs.StringVar(field, "field", "", "Output specific field only")

	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: worktree name required")
		os.Exit(1)
	}

	name := fs.Arg(0)

	if err := ShowWorktree(name, *format, *field); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleRemove(args []string) {
	fs := flag.NewFlagSet("remove", flag.ExitOnError)
	force := fs.Bool("f", false, "Skip confirmation")
	fs.BoolVar(force, "force", false, "Skip confirmation")

	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: worktree name required")
		os.Exit(1)
	}

	name := fs.Arg(0)

	if err := RemoveWorktree(name, *force); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleMCP() {
	ctx := context.Background()
	if err := StartMCPServer(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting MCP server: %v\n", err)
		os.Exit(1)
	}
}
