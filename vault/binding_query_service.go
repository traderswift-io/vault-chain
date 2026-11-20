package vault

import (
	"context"

	"google.golang.org/grpc"
)

type QueryServer interface {
	Where(context.Context, *QueryWhereRequest) (*QueryWhereResponse, error)
	MayRead(context.Context, *QueryMayReadRequest) (*QueryMayReadResponse, error)
}

func handleQueryWhere(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &QueryWhereRequest{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(QueryServer).Where(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Query/Where"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(QueryServer).Where(ctx, request.(*QueryWhereRequest))
	}
	return interceptor(ctx, message, info, handler)
}

func handleQueryMayRead(server interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	message := &QueryMayReadRequest{}
	if err := dec(message); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return server.(QueryServer).MayRead(ctx, message)
	}
	info := &grpc.UnaryServerInfo{Server: server, FullMethod: "/traderswift.vault.v1.Query/MayRead"}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return server.(QueryServer).MayRead(ctx, request.(*QueryMayReadRequest))
	}
	return interceptor(ctx, message, info, handler)
}

var QueryServiceDesc = grpc.ServiceDesc{
	ServiceName: "traderswift.vault.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Where", Handler: handleQueryWhere},
		{MethodName: "MayRead", Handler: handleQueryMayRead},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "traderswift/vault/v1/vault.proto",
}
