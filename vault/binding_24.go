package vault

import "github.com/cosmos/gogoproto/proto"

type MsgRemoveProviderResponse struct{}

func (message *MsgRemoveProviderResponse) Reset() {
	*message = MsgRemoveProviderResponse{}
}

func (message *MsgRemoveProviderResponse) String() string {
	return proto.CompactTextString(message)
}

func (*MsgRemoveProviderResponse) ProtoMessage() {}

func (*MsgRemoveProviderResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{24}
}

func init() {
	proto.RegisterType((*MsgRemoveProviderResponse)(nil), "traderswift.vault.v1.MsgRemoveProviderResponse")
}
