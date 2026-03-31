#!/usr/bin/env zsh

INSTALL_DIR="$HOME/.local/bin"

echo "Removing binary from $INSTALL_DIR..."
rm -f "$INSTALL_DIR/tasky"

echo "Uninstallation complete! (Note: this does not remove the manual PATH additions from ~/.zshrc)"

