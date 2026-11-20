package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"

	"vault-chain/vault"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

const MaximumAccountsCreatedPerBlock = 20
const firstContactAccountNumberDomain = "traderswift first contact account number "
const firstContactStoreKey = "first_contact"

type firstContactTx interface {
	sdk.Tx
	GetSigners() ([][]byte, error)
	GetSignaturesV2() ([]txsigning.SignatureV2, error)
}

type firstContactDecorator struct {
	app *TraderSwiftApp
}

func FirstContactAccountNumber(address sdk.AccAddress) uint64 {
	digest := sha256.Sum256(append([]byte(firstContactAccountNumberDomain), address...))
	return binary.BigEndian.Uint64(digest[:8]) | uint64(1)<<63
}

func accountNumberKey(number uint64) []byte {
	return binary.BigEndian.AppendUint64([]byte{1}, number)
}

func (app *TraderSwiftApp) indexGenesisAccountNumbers(ctx sdk.Context) error {
	admin, err := app.validatorSetKeeper.Admin(ctx)
	if err != nil {
		return err
	}
	address, err := sdk.AccAddressFromBech32(admin)
	if err != nil {
		return err
	}
	if !app.accountKeeper.HasAccount(ctx, address) {
		app.accountKeeper.SetAccount(ctx, app.accountKeeper.NewAccountWithAddress(ctx, address))
	}
	var failure error
	app.accountKeeper.IterateAccounts(ctx, func(account sdk.AccountI) bool {
		key := accountNumberKey(account.GetAccountNumber())
		store := app.firstContactStore.OpenKVStore(ctx)
		previous, err := store.Get(key)
		if err != nil {
			failure = err
			return true
		}
		if previous != nil && !bytes.Equal(previous, account.GetAddress()) {
			failure = vault.ErrNumberTaken
			return true
		}
		failure = store.Set(key, account.GetAddress())
		return failure != nil
	})
	return failure
}

func (decorator firstContactDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	signedTx, ok := tx.(firstContactTx)
	if !ok {
		return next(ctx, tx, simulate)
	}
	signers, err := signedTx.GetSigners()
	if err != nil {
		return ctx, err
	}
	signatures, err := signedTx.GetSignaturesV2()
	if err != nil {
		return ctx, err
	}
	if len(signers) != 1 || len(signatures) != 1 {
		return ctx, vault.ErrWrongSigner
	}
	address := sdk.AccAddress(signers[0])
	app := decorator.app
	if app.accountKeeper.HasAccount(ctx, address) {
		return next(ctx, tx, simulate)
	}
	messages := tx.GetMsgs()
	if len(messages) != 1 || signatures[0].Sequence != 0 {
		return ctx, vault.ErrWrongSigner
	}
	link, ok := messages[0].(*vault.MsgLinkAccount)
	if !ok || link.Owner != address.String() || link.Signer != link.Owner {
		return ctx, vault.ErrWrongSigner
	}
	if err := app.vaultKeeper.ValidateFirstLink(ctx, link); err != nil {
		return ctx, err
	}
	store := app.firstContactStore.OpenKVStore(ctx)
	created, err := store.Get([]byte{2})
	if err != nil {
		return ctx, err
	}
	count := uint64(0)
	if created != nil {
		if len(created) != 16 {
			return ctx, vault.ErrBadEncoding
		}
		if binary.BigEndian.Uint64(created[:8]) == uint64(ctx.BlockHeight()) {
			count = binary.BigEndian.Uint64(created[8:])
		}
	}
	if count >= MaximumAccountsCreatedPerBlock {
		return ctx, vault.ErrAccountCap
	}
	number := FirstContactAccountNumber(address)
	taken, err := store.Has(accountNumberKey(number))
	if err != nil {
		return ctx, err
	}
	if taken {
		return ctx, vault.ErrNumberTaken
	}
	account := app.accountKeeper.NewAccountWithAddress(ctx, address)
	if err := account.SetAccountNumber(number); err != nil {
		return ctx, err
	}
	app.accountKeeper.SetAccount(ctx, account)
	if err := store.Set(accountNumberKey(number), address); err != nil {
		return ctx, err
	}
	created = binary.BigEndian.AppendUint64(nil, uint64(ctx.BlockHeight()))
	created = binary.BigEndian.AppendUint64(created, count+1)
	if err := store.Set([]byte{2}, created); err != nil {
		return ctx, err
	}
	return next(ctx, tx, simulate)
}
