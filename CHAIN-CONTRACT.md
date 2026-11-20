# Vault metadata contract

chainprefix/identity.go holds the chain ID, address prefix, binary, genesis
admin and note/file scope examples. Proto package names and Go imports are
structural fork names. Content, grant witnesses, amounts and transfers are
absent from write messages.

## Authentication and first contact

The SDK ante chain verifies the signer, chain ID, account number and sequence.
Transaction envelope, auth structure, key types, multisig nesting, signature
counts and signing modes use the pinned SDK decoder and ante handlers. Both
single and multisig signers may use LEGACY_AMINO_JSON; DIRECT and other modes
are accepted only as supported by the configured SDK handler. The chain adds
no signing-mode allowlist. SDK ante verifies the threshold, selected
signatures, signer account and replay protection.

The former custom envelope, auth-size, flat-multisig, 16-key, signature-size,
fee-shape and DIRECT-only restrictions are removed for SDK compatibility.
Transaction transport and consensus allocation limits remain SDK/CometBFT
responsibilities. Chain message limits and the message whitelist remain in
force; they do not constrain the SDK's signing modes.

An unknown account's transaction must contain exactly its first LinkAccount,
with signer == owner, nonce 1, a valid hub possession proof and SDK sequence 0.
Its account number is the hub's FirstContactAccountNumber:
the first eight bytes, big endian, of SHA-256(ASCII
`traderswift first contact account number ` || raw account address), OR 2^63.
The trailing space is part of the domain. Existing accounts keep their actual
number and sequence. At most 20 first-contact accounts are created per block.
A persisted number index rejects collisions, including genesis holders. The
index is rebuilt from auth accounts at genesis. Invalid proofs, mixed messages,
wrong sequence and failed SDK signatures create no account. No transfer or fee
balance is required.

Addresses are canonical lowercase bech32, at most 90 bytes, with exactly 20
address bytes. Owner fields use the local prefix, hub fields the hub prefix.
All owner messages require signer == owner. Messages and rows are at most
32768 bytes. Unknown/duplicate fields, invalid UTF-8, nonminimal varints,
nesting over eight and more than 128 wire fields refuse before SDK decoding.
Writes use a cached context and emit vault.accepted without attributes;
SDK signer events contain no grant witnesses.

## Messages and refusal codes

| Message | Rule | Code |
|---|---|---|
| LinkAccount | owner, hub identity and possession proof; nonce must be previous nonce + 1, starting at 1 | invalid_link 10; link_replay 26 |
| SetStorageChoice | 1..3 unique phone/base/traderswift locations, 1..256-byte opaque UTF-8 pointers without controls | invalid_choice 11; location_disabled 12 |
| CommitRoot | sender's date must equal the block's UTC YYYY-MM-DD day; stored key/day are derived from block time; exactly 32 root bytes | invalid_root 13; root_exists 14 |
| Grant | 32-byte grantee commitment, scope, expiry strictly after block time and at most 366 days ahead | invalid_grant 15 |
| Revoke | owner, commitment and scope identify an existing nonrevoked grant | invalid_revoke 16 |
| RegisterFrontend | admin entry with service-key proof, ID, HTTPS endpoint and optional service signer | invalid_frontend 17 |
| RemoveFrontend | current admin removes an existing ID and releases its owner/key slot | frontend_admin_required 27; invalid_frontend 17 |
| UpdateParams | current admin sets the two storage flags | invalid_params 18; non_admin_params 21 |
| ActivateProvider | linked/authenticated owner, ID, 32-byte consensus key and 64-byte possession signature | invalid_provider 28; join_cooldown 29; inherited validator cap errors |

Distinct module refusals: wrong_signer 20, bad_frontend_signature 22,
wrong_chain_id 23, unknown_location 24 and oversize_pointer 25. Common encoding
and resource errors are bad_encoding 2, limit_exceeded 3, unauthorized 4 and
invalid_query 19. First contact adds account_cap 30 and account_number_taken 31.
SDK authentication failures retain SDK codespaces and codes.

E(domain, fields...) encodes every field, including the domain, prefixed by a
uint32 big-endian byte length. Link proof signs SHA-256(E(module + `-link-v2`,
chain ID, owner bech32, hub bech32, nonce as uint64 big endian)). The key-derived
address must equal the hub account bytes. Ed25519 signs this digest directly;
secp256k1 uses the SDK Sign/VerifySignature convention. AccountLink stores the
last nonce. An owner may change the link with a fresh proof at the next nonce;
an older signed link cannot overwrite it, including after export/import.
Hub possession does not prove a live hub ClaimOwner record or finality.

Both storage flags must be false on every InitChain, regardless of initial
height. Genesis choices must agree with their parameters; initial choices can
therefore name only phone. Disabling an enabled location later prevents new
choices without erasing existing pointers. Export preserves those records;
a launch genesis cannot import enabled flags or choices forbidden by its flags.
Snapshots requiring enabled flags need a separately reviewed restore path.

Roots are immutable and optional, one per owner per block UTC day. Past and
future sender dates both refuse. No pruning can reopen a consumed day. Client
padding and batching remain client obligations.

## Grants and queries

Commitment = SHA-256(salt || raw grantee address bytes), with exactly 32 salt
bytes and 20 address bytes, without framing or extra domain. The owner chooses
a fresh salt and shares it off chain. Neither Grant nor Revoke includes a
plain grantee or salt. Scope is 1..64 ASCII letters/digits/._:-; note/file scopes
are examples, not a closed vocabulary.

The key is (owner, commitment, scope). Regrant updates expiry and clears
revocation; revoke marks that exact grant revoked. MayRead compares exclusive
expiry against consensus time on the query event. Missing/revoked/expired or
wrong-witness grants answer no. No expiry timer runs.

Where returns the owner's link and storage choice, null if absent, as canonical
JSON with height. MayRead returns allowed and height. The query endpoint sees
the grantee and salt; transaction/state/event privacy does not cover its logs.
Owner, scope, expiry and transaction timing remain public.

## Frontends, consensus and admin

Frontend registration uses the permitted admin-entry branch. No hub-proof
adapter is installed, and self-asserted charter evidence is never trusted.
The current admin signs the transaction. ServiceSigner, if omitted, defaults
to the admin; otherwise it names the owner of the metadata slot. The service
key signs SHA-256(E(module + `-frontend-v1`, chain ID, canonical message JSON
with service_signature omitted)), binding all fields including ServiceSigner.
IDs are 1..64 ASCII letters/digits/underscore/hyphen; endpoints are 1..256-byte
HTTPS URLs without userinfo, query or fragment. IDs and service keys are unique;
each service signer has one frontend. Admin removal deletes its row, key index
and owner counter, allowing replacement. No shared 32-frontend slot cap exists.
Genesis frontend entries are explicit genesis-authority entries.

Consensus providers are separate records. ActivateProvider signs
SHA-256(E(module + `-provider-v1`, chain ID, signer bech32, provider ID, raw
consensus key)) with the consensus key. Each owner has one provider; IDs and
consensus keys are unique; at most 32 validators exist. Newcomer power is fixed
at 1. Activation calls the reference validatorset.ActivateProvider through its
installed SetStorageAdmission policy, inside the module transaction cache.
A successful newcomer activation persists consensus block time. Subsequent newcomer activation
requires at least 86400 seconds; failure changes neither membership nor time.
There is no background queue: a refused applicant resubmits on a later event.

The inherited aggregate nonfounder power must remain strictly below one third.
The stronger inherited founder-loss bound also applies: twice nonfounder power
must be less than founder power minus three times the largest founder power.
Genesis validators and providers must match bijectively by consensus key,
provider ID/moniker, power and founding flag. Administrative add/reweight and
the generic RemoveValidator route remain closed. Frontend registration grants no consensus power.
MsgRemoveProvider and MsgReplaceProviderOwner require the current admin from
SetAdmin/AcceptAdmin. Their signer field uses the same stock SDK amino multisig
transaction flow; one member of a 2-of-3 admin cannot authorize either message.
The bounded JSON CLI commands are `vault remove-provider` and
`vault replace-provider-owner`, with signer and provider_id fields; replacement
also requires new_owner. Removal calls validatorset.RemoveProvider under a
transaction-local admission authorization. A founding record becomes vacant and
retains its owner, power, founding flag and removed consensus public key. Its
power is omitted from consensus updates, active attestation powers and exported
CometBFT validators. Nonfounder removal deletes its record and owner counter,
retaining its consensus identity in the independent retired_pubkeys record.
Any removal while a founder slot is vacant is refused, including a second
removal of the same slot. There is at most one vacant founder slot.

For the founder-loss bound, a vacant founder slot counts as in recovery at its
original power. The one-third nonfounder bound uses only active voting power,
excluding vacant power. With no vacancy both inherited bounds are unchanged.
Genesis/export retain the vacancy in both module records and require matching
vacancy flags as well as the existing identity and power fields. Validator-list
queries include the vacant record and report total_power from active slots.
Genesis also checks the reserved complete set against the same founder-loss
bound as live newcomer admission, refusing a vacancy that cannot be refilled.
The public `genesis validate` (alias `validate-genesis`) and `genesis collect`
commands accept `--snapshot` for exported state, including vacancies and admin
rotation. This path retains module, cross-module and SDK validation. Without
that flag the strict initial four-equal-founder and founding-admin checks apply.
The snapshot flag retains the storage rules for every InitChain: both flags
must be false and choices must name only phone. Before exporting a recovery
snapshot after enabling storage, owners must commit phone choices and the
current admin must disable both flags through the existing messages.

Genesis module maps require auth, genutil, validatorset and vault. The consensus
module entry is optional; application export carries consensus parameters in the
separate ConsensusParams field. Unknown module names are refused. Every supplied
module still passes the JSON bounds and vault state still requires canonical
encoding, including vacant slots and retired consensus public keys.

Replacement changes only Owner and transfers the owner counter atomically; an
address already owning any provider, including the target, is refused. It may
name the owner of a vacant or live slot. ActivateProvider on an existing slot
requires that slot to be vacant, its current owner to sign, and a valid possession
proof for a new consensus public key. It restores the original founder power
and founding status without consuming the newcomer cooldown. The previous owner
cannot refill the slot after ownership changes. Activation into a live slot is
refused. No consensus-key replacement message exists.

Every removed consensus public key remains retired across subsequent recoveries
and export/import and cannot activate any provider or refill any founder slot.
Founder histories remain in their slots; nonfounder histories survive deletion
in the independent retired_pubkeys record and optional genesis field. Genesis
rejects malformed or duplicate identities across both histories and active reuse.
Each founder slot retains at most 256 retired public keys. The independent
retirement record has the existing 32768-byte record bound. Exhaustion refuses
the entire removal atomically with limit_exceeded.
All owner, provider, validator and admission changes share the transaction cache;
any later message refusal rolls back the entire transaction's module changes.
Stable module errors 32 through 39 are provider_admin_required,
provider_not_found, invalid_provider_owner, provider_owner_in_use,
provider_slot_already_vacant, provider_slot_live, provider_consensus_retired
and provider_recovery_capacity (reserved founder-loss capacity exhausted).
The validator keeper adds code 8, founder recovery in progress. Other malformed
provider fields retain invalid_provider.

This is the nonfinancial adaptation of storage admission: no storage bond,
custody, capacity or payment subsystem is copied into the metadata chain.

The genesis template admin is `chainprefix.GenesisAdmin`.
InitChain represents those same address bytes with the local prefix and ensures
an auth account exists. No multisig keys are embedded. SetAdmin/AcceptAdmin copy
the supplied two-step keeper logic with its activation gate removed. Only the
current admin may propose or replace a pending admin; only the pending admin
may accept. Acceptance clears pending and removes the old admin's authority.

## State bounds

Primary table prefixes: links 1, choices 2, roots 3, grants 4, frontends 5,
parameters 6, consensus providers 7, last admission time 8. Prefix 9 stores
per-owner/family uint64 row counters; prefix 10 indexes frontend service keys.
Keys use uint32 big-endian length framing. Counter changes and writes are atomic.
Limits per owner: one link, one storage choice of at most three locations,
65536 daily roots, 256 grant identities, one frontend, one consensus provider.
Updating an existing identity consumes no extra slot. Revoked grants keep their
slot and may be regranted. Exhausting one owner's allowance cannot reject
another owner's metadata writes. No shared row/byte budget gates those writes.

Only genesis/export scans all metadata. Ordinary writes use keyed reads;
consensus admission scans at most 32 providers and 8192 retired public keys. Export preserves nonce, grants,
provider records and last admission time. Genesis rebuilds derived counters and
frontend indexes, checking shapes, duplicate identities and per-owner limits.
JSON genesis ingress remains bounded to 16 MiB; larger operational snapshot
transport is outside this launch-genesis implementation. There is no migration
from the earlier pre-nonce/pre-provider row layout.
