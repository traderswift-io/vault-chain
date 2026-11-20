# Dependency provenance and reproducible build

Go requirements, replacements and go.sum are inherited from storage-chain.
Only the Go module declaration changes. No dependency was added or upgraded.

- go.mod SHA-256: `684746971c4cf1b74417e850ac42b2ebc2b6af291db98dbd6a27c540ef756842`
- go.sum SHA-256: `90821a92961266c98c3be662e8a58b4b7ca1410b92bc4996455ca1e48ef98e90`
- The inherited Go directive is 1.26.5; builds and checks use Go 1.27.3.
- Pinned Cosmos SDK 0.55.0, CometBFT 0.40.0 and gogoproto 1.7.2 remain unchanged.

The Makefile keeps the reference's offline toolchain/cache convention and
GOPROXY=off/GOSUMDB=off. The build uses CGO_ENABLED=0, trimpath,
buildvcs=false and an empty linker build ID. tools/archive.sh extracts the
specified git tag twice into independent temporary source paths when invoked
twice by the release check, builds each binary and creates normalized tar/gzip
archives containing the executable and its SHA256SUMS. The archive comparison
must establish identical SHA-256; source inspection alone does not establish
reproducibility. Temporary archive workspaces are removed by an EXIT trap.

Make targets are serialized with `.NOTPARALLEL`. All Go build, vet and test
invocations, including tagged archive builds, use `-p=1`; scenario execution
also uses `-parallel=1`. This limits concurrent package work without changing
dependencies, compiler optimizations or the required check coverage. It does
not cap a single compiler's memory consumption or the Go runtime's threads.

The reference supplied no archive target or deploy kit. The archive target and
isolated local-node scenario are additions. Artifacts now stay under each
checked tree's _build/ rather than a shared sibling directory.

The validatorset bindings keep pinned gocosmos regeneration and deterministic
splitting. The descriptor's Go package changes with the module rename and its
gzip bytes are normalized by tools/descriptor_bytes.py. The new metadata module
uses the existing dependency's reflection codec and a bounded stdlib-Python
schema adapter. Both binding paths have exact parity checks.

No network resolution, release signing, operational deployment or live genesis
is claimed. Go formatting, compilation,
vet, runtime scenarios, pinned generator parity and reproducibility require a
Go build run (`make check`); Python parity checks do not substitute for those gates.
