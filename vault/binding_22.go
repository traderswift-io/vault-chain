package vault

import "github.com/cosmos/gogoproto/proto"

type QueryMayReadResponse struct {
	Allowed bool   `protobuf:"varint,1,opt,name=allowed,proto3" json:"allowed,omitempty"`
	Height  uint64 `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
}

func (message *QueryMayReadResponse) Reset() {
	*message = QueryMayReadResponse{}
}

func (message *QueryMayReadResponse) String() string {
	return proto.CompactTextString(message)
}

func (*QueryMayReadResponse) ProtoMessage() {}

func (*QueryMayReadResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{22}
}

func init() {
	proto.RegisterType((*QueryMayReadResponse)(nil), "traderswift.vault.v1.QueryMayReadResponse")
}
