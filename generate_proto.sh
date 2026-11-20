#!/bin/bash
set -euo pipefail
repository_directory="$(cd "$(dirname "$0")" && pwd)"
toolchain_gopath="$HOME/.local/toolchains/gopath"
export PATH="$HOME/.local/toolchains/go/bin:$toolchain_gopath/bin:$PATH"
software_development_kit_proto="$toolchain_gopath/pkg/mod/github.com/cosmos/cosmos-sdk@v0.55.0/proto"
workspace="$repository_directory/_build/proto_workspace"
mkdir -p "$workspace/traderswift/validatorset/v1" "$workspace/cosmos/msg/v1"
install -m 644 "$repository_directory/proto/validatorset.proto" "$workspace/traderswift/validatorset/v1/validatorset.proto"
install -m 644 "$software_development_kit_proto/cosmos/msg/v1/msg.proto" "$workspace/cosmos/msg/v1/msg.proto"
cat > "$workspace/buf.gen.yaml" <<'YAML'
version: v2
plugins:
  - local: protoc-gen-gocosmos
    out: generated
    opt:
      - plugins=grpc
      - paths=source_relative
YAML
cd "$workspace"
buf generate --path traderswift/validatorset/v1/validatorset.proto
install -m 644 generated/traderswift/validatorset/v1/validatorset.pb.go validatorset.first.pb.go
buf generate --path traderswift/validatorset/v1/validatorset.proto
cmp validatorset.first.pb.go generated/traderswift/validatorset/v1/validatorset.pb.go
python3 "$repository_directory/tools/split_proto.py" "${1:-check}" "$workspace/generated" "$repository_directory"
python3 "$repository_directory/tools/module_schema.py" "${1:-check}"
printf '%s\n' 'PASS offline generated binding reproducibility and committed parity'
