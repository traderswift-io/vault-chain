package vault

import (
	"context"
	"encoding/json"

	"vault-chain/chainprefix"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type queryServer struct{ keeper *Keeper }

func NewQueryServer(keeper *Keeper) QueryServer {
	return queryServer{keeper: keeper}
}

func (server queryServer) Where(ctx context.Context, request *QueryWhereRequest) (*QueryWhereResponse, error) {
	if request == nil || !validOwner(request.Owner) {
		return nil, ErrQuery
	}
	record := WhereRecord{Link: &AccountLink{}, Choice: &StorageChoice{}}
	found, err := server.keeper.read(ctx, recordKey(linkTable, request.Owner), record.Link)
	if err != nil {
		return nil, err
	}
	if !found {
		record.Link = nil
	}
	found, err = server.keeper.read(ctx, recordKey(choiceTable, request.Owner), record.Choice)
	if err != nil {
		return nil, err
	}
	if !found {
		record.Choice = nil
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	return &QueryWhereResponse{CanonicalRecord: encoded, Height: uint64(sdk.UnwrapSDKContext(ctx).BlockHeight())}, nil
}

func (server queryServer) MayRead(ctx context.Context, request *QueryMayReadRequest) (*QueryMayReadResponse, error) {
	if request == nil || !validOwner(request.Owner) || !scopeName(request.Scope) || len(request.Salt) != 32 {
		return nil, ErrQuery
	}
	grantee, err := addressBytes(request.Grantee, chainprefix.AccountAddressPrefix)
	if err != nil {
		return nil, ErrQuery
	}
	commitment := GrantCommitment(request.Salt, grantee)
	entry := &Grant{}
	found, err := server.keeper.read(ctx, recordKey(grantTable, request.Owner, string(commitment), request.Scope), entry)
	if err != nil {
		return nil, err
	}
	context := sdk.UnwrapSDKContext(ctx)
	return &QueryMayReadResponse{Allowed: found && !entry.Revoked && entry.Expiry > blockSeconds(context), Height: uint64(context.BlockHeight())}, nil
}
