package vault

import (
	"context"

	"google.golang.org/grpc"
)

type MsgServer interface {
	LinkAccount(context.Context, *MsgLinkAccount) (*MsgLinkAccountResponse, error)
	SetStorageChoice(context.Context, *MsgSetStorageChoice) (*MsgSetStorageChoiceResponse, error)
	CommitRoot(context.Context, *MsgCommitRoot) (*MsgCommitRootResponse, error)
	Grant(context.Context, *MsgGrant) (*MsgGrantResponse, error)
	Revoke(context.Context, *MsgRevoke) (*MsgRevokeResponse, error)
	RegisterFrontend(context.Context, *MsgRegisterFrontend) (*MsgRegisterFrontendResponse, error)
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
	RemoveFrontend(context.Context, *MsgRemoveFrontend) (*MsgRemoveFrontendResponse, error)
	ActivateProvider(context.Context, *MsgActivateProvider) (*MsgActivateProviderResponse, error)
	RemoveProvider(context.Context, *MsgRemoveProvider) (*MsgRemoveProviderResponse, error)
	ReplaceProviderOwner(context.Context, *MsgReplaceProviderOwner) (*MsgReplaceProviderOwnerResponse, error)
}

func handleMsgLinkAccount(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &MsgLinkAccount{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(MsgServer).LinkAccount(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Msg/LinkAccount"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(MsgServer).LinkAccount(ctx, request.(*MsgLinkAccount))
	}
	return interceptor(ctx, message, info, handler)
}

func handleMsgSetStorageChoice(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &MsgSetStorageChoice{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(MsgServer).SetStorageChoice(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Msg/SetStorageChoice"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(MsgServer).SetStorageChoice(ctx, request.(*MsgSetStorageChoice))
	}
	return interceptor(ctx, message, info, handler)
}

func handleMsgCommitRoot(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &MsgCommitRoot{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(MsgServer).CommitRoot(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Msg/CommitRoot"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(MsgServer).CommitRoot(ctx, request.(*MsgCommitRoot))
	}
	return interceptor(ctx, message, info, handler)
}

func handleMsgGrant(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &MsgGrant{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(MsgServer).Grant(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Msg/Grant"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(MsgServer).Grant(ctx, request.(*MsgGrant))
	}
	return interceptor(ctx, message, info, handler)
}

func handleMsgRevoke(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &MsgRevoke{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(MsgServer).Revoke(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Msg/Revoke"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(MsgServer).Revoke(ctx, request.(*MsgRevoke))
	}
	return interceptor(ctx, message, info, handler)
}

func handleMsgRegisterFrontend(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &MsgRegisterFrontend{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(MsgServer).RegisterFrontend(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Msg/RegisterFrontend"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(MsgServer).RegisterFrontend(ctx, request.(*MsgRegisterFrontend))
	}
	return interceptor(ctx, message, info, handler)
}

func handleMsgUpdateParams(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &MsgUpdateParams{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(MsgServer).UpdateParams(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Msg/UpdateParams"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(MsgServer).UpdateParams(ctx, request.(*MsgUpdateParams))
	}
	return interceptor(ctx, message, info, handler)
}

func handleMsgRemoveFrontend(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &MsgRemoveFrontend{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(MsgServer).RemoveFrontend(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Msg/RemoveFrontend"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(MsgServer).RemoveFrontend(ctx, request.(*MsgRemoveFrontend))
	}
	return interceptor(ctx, message, info, handler)
}

func handleMsgActivateProvider(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &MsgActivateProvider{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(MsgServer).ActivateProvider(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Msg/ActivateProvider"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(MsgServer).ActivateProvider(ctx, request.(*MsgActivateProvider))
	}
	return interceptor(ctx, message, info, handler)
}

func handleMsgRemoveProvider(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &MsgRemoveProvider{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(MsgServer).RemoveProvider(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Msg/RemoveProvider"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(MsgServer).RemoveProvider(ctx, request.(*MsgRemoveProvider))
	}
	return interceptor(ctx, message, info, handler)
}

func handleMsgReplaceProviderOwner(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &MsgReplaceProviderOwner{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(MsgServer).ReplaceProviderOwner(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Msg/ReplaceProviderOwner"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(MsgServer).ReplaceProviderOwner(ctx, request.(*MsgReplaceProviderOwner))
	}
	return interceptor(ctx, message, info, handler)
}

var MsgServiceDesc = grpc.ServiceDesc{
	ServiceName: "traderswift.vault.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "LinkAccount", Handler: handleMsgLinkAccount},
		{MethodName: "SetStorageChoice", Handler: handleMsgSetStorageChoice},
		{MethodName: "CommitRoot", Handler: handleMsgCommitRoot},
		{MethodName: "Grant", Handler: handleMsgGrant},
		{MethodName: "Revoke", Handler: handleMsgRevoke},
		{MethodName: "RegisterFrontend", Handler: handleMsgRegisterFrontend},
		{MethodName: "UpdateParams", Handler: handleMsgUpdateParams},
		{MethodName: "RemoveFrontend", Handler: handleMsgRemoveFrontend},
		{MethodName: "ActivateProvider", Handler: handleMsgActivateProvider},
		{MethodName: "RemoveProvider", Handler: handleMsgRemoveProvider},
		{MethodName: "ReplaceProviderOwner", Handler: handleMsgReplaceProviderOwner},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "traderswift/vault/v1/vault.proto",
}
