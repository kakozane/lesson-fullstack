#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lefthook-rc.sh
source "$ROOT/scripts/lefthook-rc.sh"

exec "$@"
