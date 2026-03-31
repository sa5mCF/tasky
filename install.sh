#!/usr/bin/env zsh

INSTALL_DIR="$HOME/.local/bin"

echo "Building tasky..."
go build -o tasky ./cmd/main.go

echo "Installing tasky to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
mv tasky "$INSTALL_DIR/tasky"

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "Adding $INSTALL_DIR to PATH in ~/.zshrc..."
    echo '\n# tasky' >> "$HOME/.zshrc"
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.zshrc"
    echo "Please run 'source ~/.zshrc' or restart your terminal to apply the changes."
    source ~/.zshrc
fi

echo "Installation complete!"
