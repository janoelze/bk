#!/bin/bash
set -e

echo "Building bk..."
go build -o bk .

# Create ~/.local/bin if it doesn't exist
mkdir -p ~/.local/bin

# Remove old symlink/binary if exists
rm -f ~/.local/bin/bk

# Create symlink
ln -s "$(pwd)/bk" ~/.local/bin/bk

echo "Installed bk to ~/.local/bin/bk"

# Check if ~/.local/bin is in PATH
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo ""
    echo "Add ~/.local/bin to your PATH by adding this to your ~/.zshrc or ~/.bashrc:"
    echo ""
    echo '  export PATH="$HOME/.local/bin:$PATH"'
    echo ""
fi

# Show shell function for cd integration
echo ""
echo "To enable 'cd' on select, add this function to your ~/.zshrc or ~/.bashrc:"
echo ""
cat << 'SHELL_FUNC'
bk() {
  if [[ "$1" == "add" ]]; then
    command bk add
  else
    local dir
    dir=$(command bk)
    if [[ -n "$dir" && -d "$dir" ]]; then
      cd "$dir"
    fi
  fi
}
SHELL_FUNC
echo ""
echo "Then run: source ~/.zshrc"
