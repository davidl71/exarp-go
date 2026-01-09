# Shell Completion Scripts

This directory contains shell completion scripts for `exarp-go` to enable tab completion in bash, zsh, and fish shells.

## Installation

### Bash

**Option 1: System-wide installation (requires sudo)**
```bash
sudo cp completions/exarp-go.bash /etc/bash_completion.d/exarp-go
# or on macOS with Homebrew bash-completion:
sudo cp completions/exarp-go.bash /usr/local/etc/bash_completion.d/exarp-go
```

**Option 2: User-specific installation**
```bash
mkdir -p ~/.bash_completion.d
cp completions/exarp-go.bash ~/.bash_completion.d/
echo "source ~/.bash_completion.d/exarp-go.bash" >> ~/.bashrc
```

**Option 3: Generate dynamically (recommended)**
```bash
# Add to ~/.bashrc
eval "$(exarp-go -completion bash)"
```

### Zsh

**Option 1: Using fpath (recommended)**
```bash
# Create completions directory if it doesn't exist
mkdir -p ~/.zsh/completions

# Copy completion script
cp completions/_exarp-go ~/.zsh/completions/

# Add to ~/.zshrc
fpath=(~/.zsh/completions $fpath)
autoload -U compinit && compinit
```

**Option 2: Generate dynamically**
```bash
# Add to ~/.zshrc
eval "$(exarp-go -completion zsh)"
```

### Fish

**Option 1: Copy to fish completions directory**
```bash
# Create completions directory if it doesn't exist
mkdir -p ~/.config/fish/completions

# Copy completion script
cp completions/exarp-go.fish ~/.config/fish/completions/
```

**Option 2: Generate dynamically**
```bash
# Add to ~/.config/fish/config.fish
exarp-go -completion fish | source
```

## Features

The completion scripts provide:

- **Tool name completion**: Tab completion for all available tools when using `-tool` or `-test` flags
- **Flag completion**: Tab completion for all command-line flags
- **Shell-specific completion**: Completion for the `-completion` flag (bash, zsh, fish)

## Regenerating Completion Scripts

Completion scripts are generated dynamically from the available tools. To regenerate:

```bash
# Generate bash completion
exarp-go -completion bash > completions/exarp-go.bash

# Generate zsh completion
exarp-go -completion zsh > completions/_exarp-go

# Generate fish completion
exarp-go -completion fish > completions/exarp-go.fish
```

## Testing

After installation, restart your shell or source your configuration file, then test:

```bash
# Test tool name completion
exarp-go -tool <TAB>

# Test flag completion
exarp-go -<TAB>
```

## Troubleshooting

### Bash completion not working

1. Ensure bash-completion is installed:
   - macOS: `brew install bash-completion`
   - Linux: Usually pre-installed

2. Verify the completion script is sourced in your `~/.bashrc` or `~/.bash_profile`

3. Restart your terminal or run `source ~/.bashrc`

### Zsh completion not working

1. Ensure `compinit` is initialized in your `~/.zshrc`:
   ```zsh
   autoload -U compinit && compinit
   ```

2. Verify the completion script is in your `fpath`

3. Restart your terminal or run `source ~/.zshrc`

### Fish completion not working

1. Verify the completion script is in `~/.config/fish/completions/`

2. Restart your terminal or run `source ~/.config/fish/config.fish`

