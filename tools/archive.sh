#!/bin/bash
set -euo pipefail
repository_directory="$(cd "$(dirname "$0")/.." && pwd)"
tag="${1:?tag required}"
output="${2:?output required}"
binary_name="${3:?binary required}"
case "$output" in "$repository_directory"/_build/*.tar.gz) ;; *) echo 'FAIL archive output must be inside _build'; exit 1 ;; esac
case "$tag" in v[0-9]*) ;; *) echo 'FAIL expected version tag'; exit 1 ;; esac
case "$binary_name" in *[!a-z0-9_-]*|'') echo 'FAIL invalid binary name'; exit 1 ;; esac
commit=$(git -C "$repository_directory" rev-parse --verify "refs/tags/$tag^{commit}")
source_sha=$(git -C "$repository_directory" archive "refs/tags/$tag" | sha256sum | cut -d ' ' -f1)
workspace=$(mktemp -d "$repository_directory/_build/archive.XXXXXX")
trap 'rm -rf "$workspace"' EXIT
mkdir -p "$workspace/source" "$workspace/artifact"
git -C "$repository_directory" archive "refs/tags/$tag" | tar xf - -C "$workspace/source"
(cd "$workspace/source" && go build -p=1 -trimpath -buildvcs=false -ldflags="-buildid= -X vault-chain/cmd.GitCommit=$commit -X vault-chain/cmd.ArchiveTag=$tag -X vault-chain/cmd.ArchiveSource=$source_sha" -o "$workspace/artifact/$binary_name" .)
chmod 755 "$workspace/artifact/$binary_name"
(cd "$workspace/artifact" && sha256sum "$binary_name" > SHA256SUMS)
chmod 644 "$workspace/artifact/SHA256SUMS"
tar --sort=name --mtime=@0 --owner=0 --group=0 --numeric-owner -C "$workspace/artifact" -cf - "$binary_name" SHA256SUMS | gzip -n > "$output"
printf '%s\n' 'PASS archive build'
