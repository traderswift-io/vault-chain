package vault

import "github.com/cosmos/gogoproto/proto"

type MsgRevokeResponse struct{}

func (message *MsgRevokeResponse) Reset() {
	*message = MsgRevokeResponse{}
}

func (message *MsgRevokeResponse) String() string {
	return proto.CompactTextString(message)
}

func (*MsgRevokeResponse) ProtoMessage() {}

func (*MsgRevokeResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{10}
}

func init() {
	proto.RegisterType((*MsgRevokeResponse)(nil), "traderswift.vault.v1.MsgRevokeResponse")
}
