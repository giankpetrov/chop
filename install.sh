#!/bin/sh
set -e

REPO="AgusRdz/chop"
INSTALL_DIR="${CHOP_INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Linux*)  OS="linux" ;;
  Darwin*) OS="darwin" ;;
  MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
  *) echo "unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

EXT=""
if [ "$OS" = "windows" ]; then
  EXT=".exe"
fi

BINARY="chop-${OS}-${ARCH}${EXT}"

# Get latest version
if [ -z "$CHOP_VERSION" ]; then
  CHOP_VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"//;s/".*//')
fi

if [ -z "$CHOP_VERSION" ]; then
  echo "failed to determine latest version" >&2
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${CHOP_VERSION}/${BINARY}"

echo "installing chop ${CHOP_VERSION} (${OS}/${ARCH})..."

mkdir -p "$INSTALL_DIR"
curl -fsSL "$URL" -o "${INSTALL_DIR}/chop${EXT}"
chmod +x "${INSTALL_DIR}/chop${EXT}"

echo "installed chop to ${INSTALL_DIR}/chop${EXT}"
echo ""

# Check if install dir is in PATH
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    # Detect shell config file
    SHELL_NAME="$(basename "${SHELL:-}")"
    case "$SHELL_NAME" in
      zsh)  SHELL_RC="$HOME/.zshrc" ;;
      bash) SHELL_RC="$HOME/.bashrc" ;;
      *)    SHELL_RC="" ;;
    esac

    PATH_LINE="export PATH=\"${INSTALL_DIR}:\$PATH\""

    if [ -n "$SHELL_RC" ]; then
      # Only add if not already present
      if ! grep -qF "$INSTALL_DIR" "$SHELL_RC" 2>/dev/null; then
        printf '\n# chop\n%s\n' "$PATH_LINE" >> "$SHELL_RC"
        echo "Added ${INSTALL_DIR} to PATH in $SHELL_RC"
        echo "Reload your shell with: source $SHELL_RC"
      fi
    else
      echo "NOTE: ${INSTALL_DIR} is not in your PATH."
      echo "Add this line to your shell config file:"
      echo "  $PATH_LINE"
    fi
    echo ""
    ;;
esac

echo "Next steps:"
echo ""
echo "  # Use directly with any command:"
echo "  chop git status"
echo "  chop docker ps"
echo ""
echo "  # Claude Code hook (auto-rewrite Bash tool calls):"
echo "  chop init --global"
echo "  chop init --status    # check if installed"
