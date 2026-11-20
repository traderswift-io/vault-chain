#!/bin/bash
set -eu
cd "$(dirname "$0")/.."
export PATH="$HOME/.local/toolchains/go/bin:$PATH"
make check-twice
