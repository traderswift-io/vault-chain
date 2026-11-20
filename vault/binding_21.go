package vault

import "github.com/cosmos/gogoproto/proto"

type QueryMayReadRequest struct {
	Owner   string `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Grantee string `protobuf:"bytes,2,opt,name=grantee,proto3" json:"grantee,omitempty"`
	Salt    []byte `protobuf:"bytes,3,opt,name=salt,proto3" json:"salt,omitempty"`
	Scope   string `protobuf:"bytes,4,opt,name=scope,proto3" json:"scope,omitempty"`
}

func (message *QueryMayReadRequest) Reset() {
	*message = QueryMayReadRequest{}
}

func (message *QueryMayReadRequest) String() string {
	return proto.CompactTextString(message)
}

func (*QueryMayReadRequest) ProtoMessage() {}

func (*QueryMayReadRequest) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{21}
}

func init() {
	proto.RegisterType((*QueryMayReadRequest)(nil), "traderswift.vault.v1.QueryMayReadRequest")
}
