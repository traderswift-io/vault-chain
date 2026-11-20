package vault

import "github.com/cosmos/gogoproto/proto"

type MsgReplaceProviderOwnerResponse struct{}

func (message *MsgReplaceProviderOwnerResponse) Reset() {
	*message = MsgReplaceProviderOwnerResponse{}
}

func (message *MsgReplaceProviderOwnerResponse) String() string {
	return proto.CompactTextString(message)
}

func (*MsgReplaceProviderOwnerResponse) ProtoMessage() {}

func (*MsgReplaceProviderOwnerResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{26}
}

func init() {
	proto.RegisterType((*MsgReplaceProviderOwnerResponse)(nil), "traderswift.vault.v1.MsgReplaceProviderOwnerResponse")
}
