go_environment = PATH=$(HOME)/.local/toolchains/go/bin:$(PATH) GOPATH=$(HOME)/.local/toolchains/gopath GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off CGO_ENABLED=0
artifact_directory := $(abspath _build)
binary_name := $(shell sed -n 's/.*BinaryName *= *"\([^"]*\)".*/\1/p' chainprefix/identity.go)
archive_tag ?= v0.0.0-check
archive_output ?= $(artifact_directory)/release.tar.gz

.NOTPARALLEL:

.PHONY: build check check-twice prepare-build check-parsers check-format check-scenarios check-layout check-proto proto check-schema archive

check: check-parsers check-schema check-format check-layout build
	@$(go_environment) go vet -p=1 ./...
	@echo 'PASS go vet'
	@$(MAKE) check-scenarios check-proto
	@echo 'PASS $(binary_name) check'

check-twice:
	@$(MAKE) check
	@$(MAKE) check
	@echo 'PASS $(binary_name) check twice'

prepare-build:
	@mkdir -p $(artifact_directory)

check-parsers:
	@python3 tools/check_parsers.py

check-schema:
	@python3 tools/module_schema.py

check-format: prepare-build
	@files=$$($(go_environment) gofmt -l $$(find . -type d \( -name .git -o -name build -o -name _build \) -prune -o -type f -name '*.go' -print)) || exit $$?; test -z "$$files" || { printf '%s\n' "$$files"; $(go_environment) gofmt -d $$files; exit 1; }
	@echo 'PASS gofmt'

check-layout:
	@python3 tools/check_layout.py

build: prepare-build
	@$(go_environment) go build -p=1 -trimpath -buildvcs=false -ldflags='-buildid= -X vault-chain/cmd.GitCommit=$(shell git rev-parse HEAD)' -o $(artifact_directory)/$(binary_name) .
	@echo 'PASS go build'

check-scenarios: prepare-build
	@$(go_environment) go test -p=1 -parallel=1 -count=1 -v ./... > $(artifact_directory)/scenarios.log 2>&1 || { cat $(artifact_directory)/scenarios.log; exit 1; }
	@cat $(artifact_directory)/scenarios.log
	@echo 'PASS scenarios'

archive: prepare-build
	@$(go_environment) bash tools/archive.sh '$(archive_tag)' '$(archive_output)' '$(binary_name)'

check-proto:
	@bash generate_proto.sh check

proto:
	bash generate_proto.sh write

.PHONY: scenario-binary
scenario-binary: prepare-build
	@$(go_environment) go test -p=1 -c -o $(artifact_directory)/chain-scenarios .
