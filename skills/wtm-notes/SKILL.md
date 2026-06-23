---
name: wtm-notes
description: >
  Manage worktree notes using the wtm CLI. Triggers on: add note, show note, edit note,
  remove note, delete note, worktree note, worktree notes, annotate worktree, attach note,
  view note, note management.
---

# Worktree Notes with wtm

Attach freeform text notes to worktrees for recording context, decisions, or task references.

## Commands

### wtm notes add

```bash
wtm notes add <worktree> -m "message"
wtm notes add <worktree>           # opens $EDITOR
wtm notes add <worktree> -m "msg" -f  # overwrite existing note
```

### wtm notes show

```bash
wtm notes show <worktree>
```

Prints the note to stdout. Errors if no note exists.

### wtm notes edit

```bash
wtm notes edit <worktree>
```

Opens the note in `$EDITOR`. Creates the file if it doesn't exist.

### wtm notes remove

```bash
wtm notes remove <worktree>
```

Deletes the note. Errors if no note exists.

## Integration with other commands

Notes can also be managed through worktree commands (see `wtm-worktree` skill):

- Attach note at creation: `wtm add <name> -m "message"`
- View note in worktree info: `wtm show <name>` (note appears in pretty output)
- Output only the note: `wtm show <name> -f note`

## Storage

Notes are stored in `.git/worktrees/<name>/wtm-notes` (linked worktrees) or `.git/wtm-notes` (main worktree) and are automatically cleaned up when the worktree is removed.
