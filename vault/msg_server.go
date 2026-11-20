package vault

import "context"

type msgServer struct{ keeper *Keeper }

func NewMsgServer(keeper *Keeper) MsgServer {
	return msgServer{keeper: keeper}
}

func (server msgServer) LinkAccount(ctx context.Context, message *MsgLinkAccount) (*MsgLinkAccountResponse, error) {
	if message == nil {
		return nil, ErrBadEncoding
	}
	if err := server.keeper.Execute(ctx, message); err != nil {
		return nil, err
	}
	return &MsgLinkAccountResponse{}, nil
}

func (server msgServer) SetStorageChoice(ctx context.Context, message *MsgSetStorageChoice) (*MsgSetStorageChoiceResponse, error) {
	if message == nil {
		return nil, ErrBadEncoding
	}
	if err := server.keeper.Execute(ctx, message); err != nil {
		return nil, err
	}
	return &MsgSetStorageChoiceResponse{}, nil
}

func (server msgServer) CommitRoot(ctx context.Context, message *MsgCommitRoot) (*MsgCommitRootResponse, error) {
	if message == nil {
		return nil, ErrBadEncoding
	}
	if err := server.keeper.Execute(ctx, message); err != nil {
		return nil, err
	}
	return &MsgCommitRootResponse{}, nil
}

func (server msgServer) Grant(ctx context.Context, message *MsgGrant) (*MsgGrantResponse, error) {
	if message == nil {
		return nil, ErrBadEncoding
	}
	if err := server.keeper.Execute(ctx, message); err != nil {
		return nil, err
	}
	return &MsgGrantResponse{}, nil
}

func (server msgServer) Revoke(ctx context.Context, message *MsgRevoke) (*MsgRevokeResponse, error) {
	if message == nil {
		return nil, ErrBadEncoding
	}
	if err := server.keeper.Execute(ctx, message); err != nil {
		return nil, err
	}
	return &MsgRevokeResponse{}, nil
}

func (server msgServer) RegisterFrontend(ctx context.Context, message *MsgRegisterFrontend) (*MsgRegisterFrontendResponse, error) {
	if message == nil {
		return nil, ErrBadEncoding
	}
	if err := server.keeper.Execute(ctx, message); err != nil {
		return nil, err
	}
	return &MsgRegisterFrontendResponse{}, nil
}

func (server msgServer) UpdateParams(ctx context.Context, message *MsgUpdateParams) (*MsgUpdateParamsResponse, error) {
	if message == nil {
		return nil, ErrBadEncoding
	}
	if err := server.keeper.Execute(ctx, message); err != nil {
		return nil, err
	}
	return &MsgUpdateParamsResponse{}, nil
}

func (server msgServer) RemoveFrontend(ctx context.Context, message *MsgRemoveFrontend) (*MsgRemoveFrontendResponse, error) {
	if message == nil {
		return nil, ErrBadEncoding
	}
	if err := server.keeper.Execute(ctx, message); err != nil {
		return nil, err
	}
	return &MsgRemoveFrontendResponse{}, nil
}

func (server msgServer) ActivateProvider(ctx context.Context, message *MsgActivateProvider) (*MsgActivateProviderResponse, error) {
	if message == nil {
		return nil, ErrBadEncoding
	}
	if err := server.keeper.Execute(ctx, message); err != nil {
		return nil, err
	}
	return &MsgActivateProviderResponse{}, nil
}

func (server msgServer) RemoveProvider(ctx context.Context, message *MsgRemoveProvider) (*MsgRemoveProviderResponse, error) {
	if message == nil {
		return nil, ErrBadEncoding
	}
	if err := server.keeper.Execute(ctx, message); err != nil {
		return nil, err
	}
	return &MsgRemoveProviderResponse{}, nil
}

func (server msgServer) ReplaceProviderOwner(ctx context.Context, message *MsgReplaceProviderOwner) (*MsgReplaceProviderOwnerResponse, error) {
	if message == nil {
		return nil, ErrBadEncoding
	}
	if err := server.keeper.Execute(ctx, message); err != nil {
		return nil, err
	}
	return &MsgReplaceProviderOwnerResponse{}, nil
}
