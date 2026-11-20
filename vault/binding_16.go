package vault

import "github.com/cosmos/gogoproto/proto"

type MsgRemoveFrontendResponse struct{}

func (message *MsgRemoveFrontendResponse) Reset() {
	*message = MsgRemoveFrontendResponse{}
}

func (message *MsgRemoveFrontendResponse) String() string {
	return proto.CompactTextString(message)
}

func (*MsgRemoveFrontendResponse) ProtoMessage() {}

func (*MsgRemoveFrontendResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{16}
}

func init() {
	proto.RegisterType((*MsgRemoveFrontendResponse)(nil), "traderswift.vault.v1.MsgRemoveFrontendResponse")
}
