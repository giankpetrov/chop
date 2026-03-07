#!/bin/sh
set -e

OLD_DIR="$HOME/bin"
NEW_DIR="$HOME/.local/bin"
BINARY="chop"

echo "chop migration: $OLD_DIR -> $NEW_DIR"
echo ""

# Check if old binary exists
if [ ! -f "$OLD_DIR/$BINARY" ]; then
  echo "chop not found in $OLD_DIR — nothing to migrate."
  exit 0
fi

# Move binary
mkdir -p "$NEW_DIR"
mv "$OLD_DIR/$BINARY" "$NEW_DIR/$BINARY"
echo "moved: $OLD_DIR/$BINARY -> $NEW_DIR/$BINARY"

# Detect shell config
SHELL_NAME="$(basename "${SHELL:-}")"
case "$SHELL_NAME" in
  zsh)  SHELL_RC="$HOME/.zshrc" ;;
  bash) SHELL_RC="$HOME/.bashrc" ;;
  *)    SHELL_RC="" ;;
esac

if [ -n "$SHELL_RC" ] && [ -f "$SHELL_RC" ]; then
  # Remove old PATH entry (the comment line and the export line added by installer)
  if grep -qF "$OLD_DIR" "$SHELL_RC"; then
    grep -v "# chop" "$SHELL_RC" | grep -v "$OLD_DIR" > "$SHELL_RC.chop_migrate_tmp"
    mv "$SHELL_RC.chop_migrate_tmp" "$SHELL_RC"
    echo "removed $OLD_DIR from PATH in $SHELL_RC"
  fi

  # Add new PATH entry if not already present
  if ! grep -qF "$NEW_DIR" "$SHELL_RC"; then
    printf '\n# chop\nexport PATH="%s:$PATH"\n' "$NEW_DIR" >> "$SHELL_RC"
    echo "added $NEW_DIR to PATH in $SHELL_RC"
  fi

  echo ""
  echo "reload your shell:"
  echo "  source $SHELL_RC"
else
  echo ""
  echo "could not detect shell config — add this manually to your shell config:"
  echo "  export PATH=\"$NEW_DIR:\$PATH\""
fi

echo ""
echo "migration complete."
