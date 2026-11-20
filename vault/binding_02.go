package vault

import "github.com/cosmos/gogoproto/proto"

type MsgLinkAccountResponse struct{}

func (message *MsgLinkAccountResponse) Reset() {
	*message = MsgLinkAccountResponse{}
}

func (message *MsgLinkAccountResponse) String() string {
	return proto.CompactTextString(message)
}

func (*MsgLinkAccountResponse) ProtoMessage() {}

func (*MsgLinkAccountResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{2}
}

func init() {
	proto.RegisterType((*MsgLinkAccountResponse)(nil), "traderswift.vault.v1.MsgLinkAccountResponse")
}
