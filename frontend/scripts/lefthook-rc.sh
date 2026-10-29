#!/usr/bin/env bash

# Ensure git hooks (including GUI editor commits) use the Node version
# from .node-version, not whatever happens to be first in PATH.

resolve_nvm_dir() {
  if [ -n "${NVM_DIR:-}" ] && [ -d "$NVM_DIR" ]; then
    printf '%s' "$NVM_DIR"
    return 0
  fi

  if [ -d "$HOME/.nvm" ]; then
    printf '%s' "$HOME/.nvm"
    return 0
  fi

  if [ -d "$HOME/nvm" ]; then
    printf '%s' "$HOME/nvm"
    return 0
  fi
}

setup_node_path() {
  local root node_version nvm_dir node_bin

  root="$(git rev-parse --show-toplevel 2>/dev/null)" || return 0
  node_version="$(tr -d '[:space:]' <"$root/.node-version" 2>/dev/null)" || return 0
  [ -n "$node_version" ] || return 0

  # nvm
  nvm_dir="$(resolve_nvm_dir)" || true
  if [ -n "$nvm_dir" ]; then
    node_bin="$nvm_dir/versions/node/v${node_version}/bin/node"
    if [ -x "$node_bin" ]; then
      export PATH="$(dirname "$node_bin"):$PATH"
      return 0
    fi
  fi

  # fnm
  node_bin="$HOME/.local/share/fnm/node-versions/v${node_version}/installation/bin/node"
  if [ -x "$node_bin" ]; then
    export PATH="$(dirname "$node_bin"):$PATH"
    return 0
  fi

  # asdf
  node_bin="$HOME/.asdf/installs/nodejs/${node_version}/bin/node"
  if [ -x "$node_bin" ]; then
    export PATH="$(dirname "$node_bin"):$PATH"
    return 0
  fi

  # mise: only override when the active version matches .node-version
  if command -v mise >/dev/null 2>&1; then
    local mise_node mise_version
    mise_node="$(mise which node 2>/dev/null)" || return 0
    mise_version="$("$mise_node" -v 2>/dev/null | sed 's/^v//')" || return 0
    if [ "$mise_version" = "$node_version" ]; then
      export PATH="$(dirname "$mise_node"):$PATH"
    fi
  fi
}

setup_node_path
