# vault-chain

vault-chain runs `vault-1`, the TraderSwift chain that records where a person's notes and files are stored and who may read them; the content itself stays off chain. Four founding validators keep it, all run by TraderSwift today. It has no token and no balances.

**Just want to see it? Start here:** [what the chain is](https://github.com/traderswift-io/vault-chain#vault-chain), then [the chain contract](https://github.com/traderswift-io/vault-chain/blob/main/CHAIN-CONTRACT.md): every message, query and refusal, and [the four founding validators](https://github.com/traderswift-io/vault-chain/blob/main/deploy/README.md#vault-1-founders-launched-19-nov-2026-v110-at-a2daf91): node IDs, owners and genesis digest, readable in the browser. No install, no keys, no sign-up.

`vault-1` is a Cosmos SDK metadata chain for notes and files. Its module stores
account links, bounded opaque storage pointers, optional daily roots, private
read-grant commitments and frontend registrations. Content and grant salts stay
off chain. There is no bank, token, staking, mint, payment or balance API.

`phone` is always selectable. `base` and `traderswift` start disabled and require
separate booleans set by the current validatorset admin. A grant carries only
SHA-256(salt || raw grantee address); the read query supplies the witness privately
to the query endpoint. Queries are Where and MayRead. See CHAIN-CONTRACT.md for
wire limits, refusal codes, authorization and privacy boundaries.

Run `make check` for parser inventory, schema parity, formatting, source layout,
build, vet, scenarios and pinned validatorset binding regeneration. Run
`make check-twice` for repeated checks. Builds run offline (`GOPROXY=off`):
the Makefile expects Go, the module cache and the pinned generator tools under
`$HOME/.local/toolchains`; `tools/run_check.sh` runs `make check-twice` with that path. No dependency was added. Artifacts are in `_build/`.

Make targets run serially, including when invoked with `-j`. Go build, vet and
test commands use `-p=1`, and scenarios use `-parallel=1`, to reduce simultaneous
compiler and test processes on the build machine. Archive builds use the same
package limit. These limits reduce peak resource contention; they are not a
hard memory limit or proof of why an unrelated process exited. The separate
four-node network check remains mandatory.

`make archive archive_tag=v0.0.0-check` builds from the tagged git archive, with
trimmed paths, no VCS stamping, an empty build ID and deterministic tar/gzip
metadata. The release check builds the archive twice, compares SHA-256,
then runs four-node operational qualification in separate fresh trees. All four
founders must sign, height must exceed 20, and stopping one founder must leave
the other three producing blocks.

`TestLocalRoundTrip` launches one actual CometBFT validator with temporary keys,
on two free loopback ports selected from 47000–47999, when explicitly given a fresh
`-local-chain-directory .../_build/roundtrip` and distinct `-local-chain-p2p-port`
and `-local-chain-rpc-port` values. The directory must belong to the current tree;
ports outside the reserved range are refused before node initialization. The
check takes one listener snapshot immediately before launch, skips occupied
ports without signalling their owners, records its choices in
`_build/roundtrip.ports`, and checks those ports again after its child exits.
A concurrent bind can still make startup fail; it never takes over a listener.
It broadcasts SDK-signed messages,
queries committed state, verifies revocation, exports state, scans events for
private witness material and stops the node. It waits on consensus events; no
polling loop drives application logic. Before the first broadcast, it subscribes
to block events and checks ABCI's committed height once, then waits for the
initial commit if necessary. The transaction owner is absent from genesis: its first signed AccountLink
creates its SDK account, then the same owner chooses, commits, grants, queries
and revokes. A test-only module initializes four synthetic founding records
through production validation, but returns only the local validator's initial
consensus update. The single-node fixture does not qualify membership liveness.
Production genesis retains the inherited aggregate and founder-loss power caps.

The binary `vaultd` provides the operational root copied from the reference,
with init, start, genesis assembly/validation, keys, module queries and
transactions, export and version commands. See deploy/README.md for the founding
procedure, configuration overlays and service template. The existing app,
consensus rules and offline qualification scenarios remain unchanged.

Provider activation is wired to the inherited validator-set admission path,
with consensus-key possession, fixed newcomer power, the one-third and
founder-loss caps, and one successful add per 24 hours. Frontends have no voting
power. Registration and removal of frontend metadata require the current admin;
there is no unverified hub-charter evidence acceptance path.

The genesis admin defaults to `chainprefix.GenesisAdmin`.
Its address bytes are represented with the vault prefix in local auth state;
no key or signer membership is included. InitChain creates the admin auth
account if absent. Two-step SetAdmin/AcceptAdmin works from genesis with no
height gate. The decoder retains the chain message checks and transaction size
limit. Sign modes, public keys and signature structures use the stock SDK decoder and ante
handler, including amino-json for single signers and threshold multisig.

See MODULE.md for implementation boundaries and CHAIN-CONTRACT.md for the
nonce, per-owner limits, exact signing transcripts and admission rules.
