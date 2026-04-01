#!/usr/bin/env bash
set -euo pipefail

[ $(id -u) -eq 0 ] || exec "$@"
exec sudo -u "#$SUDO_UID" -g "#$SUDO_GID" "$@"
