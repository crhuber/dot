# dot - Dotfiles Management Tool

A command-line tool for managing dotfiles using symbolic links and profiles. Written in Go.

## Features

- **Profile-based configuration**: Organize dotfiles for different environments (general, work, etc.)
- **Symbolic link management**: Automatically create, check, and clean symbolic links
- **Repository cloning**: Clone dotfiles repositories from remote URLs
- **Repository updates**: Update your dotfiles with a single command
- **Quick access**: Open dotfiles directory in your file manager
- **Backup functionality**: Automatically backup existing files before linking
- **Dry-run support**: Preview changes before applying them
- **Environment variable support**: Override default paths with `$DOT_DIR`

## Installation

1. Download and add binary to $PATH from https://github.com/crhuber/dot/releases

Or

2. Use [kelp](https://github.com/crhuber/dot)

```bash
kelp add crhuber/dot --install
```

## Quick Start

1. **Clone your dotfiles repository:**
   ```bash
   dot clone https://github.com/yourusername/dotfiles.git
   ```

2. **Create symbolic links:**
   ```bash
   dot link
   ```

3. **Check link status:**
   ```bash
   dot check
   ```

## Commands

### `dot clone <repository-url>`
Clone a dotfiles repository to `~/.dotfiles` (or `$DOT_DIR`).

```bash
dot clone https://github.com/yourusername/dotfiles.git
dot clone git@github.com:yourusername/dotfiles.git
```

### `dot link [--profile <profiles>] [--dry-run]`
Create symbolic links based on the `.mappings` file.

```bash
# Link default profile
dot link

# Link specific profiles
dot link --profile work
dot link --profile general,work

# Preview changes without applying
dot link --dry-run
```

### `dot check [--profile <profiles>]`
Verify that symbolic links exist and point to correct sources.

```bash
# Check default profile
dot check

# Check specific profiles
dot check --profile work
```

### `dot clean [--profile <profiles>]`
Remove symbolic links defined in profiles.

```bash
# Clean default profile
dot clean

# Clean specific profiles
dot clean --profile work
```

### `dot root`
Print the dotfiles repository path.

```bash
dot root
# Output: /Users/username/.dotfiles
```

### `dot update`
Update the dotfiles repository by running git pull.

```bash
dot update
```

This command changes to your dotfiles directory and runs `git pull` to fetch and merge the latest changes from the remote repository.

### `dot open`
Open the dotfiles directory in your system's file manager.

```bash
dot open
```

This command opens the dotfiles directory in:
- Finder on macOS
- File Explorer on Windows  
- Default file manager on Linux (using xdg-open)

## Configuration

### `.mappings` File Format

The `.mappings` file in your dotfiles repository defines how files are linked. It uses TOML format:

```toml
[general]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig" = "~/.gitconfig"
"zsh/.zshrc" = "~/.zshrc"

[work]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig-work" = "~/.gitconfig"
```

- **Source paths** are relative to your dotfiles repository
- **Target paths** use `~` for your home directory
- **`[general]` profile** is required and used as default
- **Profile precedence**: Later profiles override earlier ones

### Environment Variables

- **`$DOT_DIR`**: Override the default repository location (`~/.dotfiles`)

```bash
export DOT_DIR="/custom/path"
dot clone https://github.com/yourusername/dotfiles.git
# Clones to /custom/path instead of ~/.dotfiles
```

## Examples

### Basic Workflow

```bash
# Clone your dotfiles
dot clone https://github.com/yourusername/dotfiles.git

# Link all files from general profile
dot link

# Check that everything is linked correctly
dot check

# Preview work profile changes
dot link --profile general,work --dry-run

# Apply work profile
dot link --profile general,work
```

### Example `.mappings` File

```toml
[general]
"vim/.vimrc" = "~/.vimrc"
"vim/.vim" = "~/.vim"
"git/.gitconfig" = "~/.gitconfig"
"zsh/.zshrc" = "~/.zshrc"
"tmux/.tmux.conf" = "~/.tmux.conf"

[work]
"git/.gitconfig-work" = "~/.gitconfig"
"ssh/work_config" = "~/.ssh/config"

[minimal]
"vim/.vimrc" = "~/.vimrc"
"git/.gitconfig" = "~/.gitconfig"
```

### Example Repository Structure

```
~/.dotfiles/
├── .mappings
├── vim/
│   ├── .vimrc
│   └── .vim/
├── git/
│   ├── .gitconfig
│   └── .gitconfig-work
├── zsh/
│   └── .zshrc
└── tmux/
    └── .tmux.conf
```

## Dependencies

- **Go 1.24+**
- **git**: Required for cloning repositories

## Development

```bash
# Run tests
go test ./...

# Build
go build -o dot cmd/dot/main.go

# Install locally
go install ./cmd/dot
```

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Inspiration

This tool is inspired by [ubnt-intrepid/dot](https://github.com/ubnt-intrepid/dot) but implements additional features like profile precedence, dry-run mode, and improved error handling.
