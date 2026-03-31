# Tasky

A simple, fast, and terminal-based task manager built in Go.

## Overview

Tasky allows you to track and manage your to-do lists straight from your terminal. It stores tasks in a local `.dataTodo.json` file and provides a colorful table view to easily check task statuses such as pending, doing, or completed.

## Installation

### Prerequisites
- [Go](https://go.dev/doc/install) installed on your system.
- `zsh` terminal (the installer scripts are built for zsh).

### Quick Install

You can easily build and install tasky to your `$HOME/.local/bin` directory by running the included install script:

```bash
chmod +x install.sh
./install.sh
```
*Note: Make sure to restart your terminal or run `source ~/.zshrc` if the script adds the directory to your `$PATH`.*

### Uninstall

To uninstall Tasky:

```bash
chmod +x uninstall.sh
./uninstall.sh
```

## Usage

You can interact with Tasky by using the following flags:

| Flag | Description | Example |
| :--- | :--- | :--- |
| **No Flag** | Lists all tasks (defaults to `-list`) | `tasky` |
| `-list` | Lists all tasks | `tasky -list` |
| `-add` | Adds a new task | `tasky -add "Buy groceries"` |
| `-doing` | Marks a task as "Doing" (by ID) | `tasky -doing 1` |
| `-complete` | Marks a task as "Done" (by ID) | `tasky -complete 1` |
| `-delete` | Deletes a task from the list (by ID) | `tasky -delete 1` |

*(The ID is the number shown in the `#` column when you list your tasks).*

## Building manually with Make

You can also use the included `Makefile` to quickly run commands under the hood:

- `make list` - Lists all tasks
- `make add` - Runs the add command
- `make complete` - Runs the complete command

## Tech Stack
- [Go](https://go.dev/)
- [simpletable](https://github.com/alexeyco/simpletable) for terminal table formatting.
