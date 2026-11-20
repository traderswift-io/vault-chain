package vault

import "github.com/cosmos/gogoproto/proto"

type QueryWhereResponse struct {
	CanonicalRecord []byte `protobuf:"bytes,1,opt,name=canonical_record,json=canonicalRecord,proto3" json:"canonical_record,omitempty"`
	Height          uint64 `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
}

func (message *QueryWhereResponse) Reset() {
	*message = QueryWhereResponse{}
}

func (message *QueryWhereResponse) String() string {
	return proto.CompactTextString(message)
}

func (*QueryWhereResponse) ProtoMessage() {}

func (*QueryWhereResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{20}
}

func init() {
	proto.RegisterType((*QueryWhereResponse)(nil), "traderswift.vault.v1.QueryWhereResponse")
}
