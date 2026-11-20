# Vault module implementation

The app retains auth, consensus, genutil and validatorset, with the vault
metadata module replacing storage. Nine SDK message methods and two queries
are described in CHAIN-CONTRACT.md. No Go dependency was added.

ante_first_contact.go ports the supplied hub account-number domain, zero
sequence, account collision and 20-per-block rules. The allowed first contact
is narrowed to exactly one owner AccountLink and validates its hub proof before
creating auth state. The first_contact store holds the block counter and number
index, rebuilt from auth genesis. The genesis admin also receives an auth
account if absent. The SDK performs final signature checks and rollback.

vault/owner_budget.go maintains per-owner, per-family counts without a shared
metadata capacity gate. vault/link_frontend.go handles nonce links and the
minimal admin-entry frontend registry, including removal and index release.
vault/provider_admission.go installs the reference validator admission callback
and exposes event-driven consensus-key possession activation, fixed power and
24-hour cooldown. No monetary storage protocol is imported. The inherited
aggregate and founder-loss caps stay in validatorset. Genesis checks the
validator/provider bijection. No timer, polling, Redis or remote service is used.

Two-step admin rotation is the supplied implementation with the activation
check removed. The default hub multisig address is converted to the same local
address bytes only at genesis. Bounded transaction decoding permits one direct
signer, including a flat threshold multisig; SDK verification remains mandatory.
The signer bitmap, component keys and aggregate signatures have explicit bounds.

records.go defines canonical rows; handlers.go uses cached atomic writes;
validation.go defines transcripts and witnesses; genesis.go validates and
imports/exports rows; query_server.go performs keyed reads. The local module
has no block hook; grant expiry is evaluated only on a query event.

proto/vault.proto is the source for tools/module_schema.py, which emits
registered descriptors, reflection bindings and typed services. Exact parity is
checked by make check-schema. Validatorset retains pinned gocosmos regeneration
and splitting, checked twice by make check-proto. PARSER-INVENTORY.md records
bounded codec sites. Go formatting, build, vet and runtime checks need the Go toolchain
(`make check`); the Python tools alone do not cover them.

R1 scenarios cover the real SDK first-contact path, collisions, invalid proofs,
mixed transactions, block account caps, nonce replay, owner quotas, root dates,
frontend admission/removal, genesis validation, signed admin rotation, a
throwaway multisig admin and provider cooldown/cap refusals. The local CometBFT
scenario starts a fresh non-genesis owner and exercises both queries and privacy.
The single-validator harness uses a test-only module override: production
validation initializes four synthetic founders, while the override returns only
one initial consensus update. No exported production local-check initializer
exists. This fixture is not a production genesis or a membership-liveness proof.
validatorset/local_check.go holds only the private validated-genesis writer;
its sole caller is InitializeFromGenesis after full genesis validation.

chainprefix/identity.go centralizes vault-specific values. Go module/import
paths, proto namespace/filenames, vault/ and generated bindings are structural
names for the next fork; accompanying prose names the current chain.
