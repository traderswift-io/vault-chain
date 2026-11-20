package vault

import "github.com/cosmos/gogoproto/proto"

type QueryWhereRequest struct {
	Owner string `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
}

func (message *QueryWhereRequest) Reset() {
	*message = QueryWhereRequest{}
}

func (message *QueryWhereRequest) String() string {
	return proto.CompactTextString(message)
}

func (*QueryWhereRequest) ProtoMessage() {}

func (*QueryWhereRequest) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{19}
}

func init() {
	proto.RegisterType((*QueryWhereRequest)(nil), "traderswift.vault.v1.QueryWhereRequest")
}
