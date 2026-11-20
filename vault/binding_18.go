package vault

import "github.com/cosmos/gogoproto/proto"

type MsgActivateProviderResponse struct{}

func (message *MsgActivateProviderResponse) Reset() {
	*message = MsgActivateProviderResponse{}
}

func (message *MsgActivateProviderResponse) String() string {
	return proto.CompactTextString(message)
}

func (*MsgActivateProviderResponse) ProtoMessage() {}

func (*MsgActivateProviderResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{18}
}

func init() {
	proto.RegisterType((*MsgActivateProviderResponse)(nil), "traderswift.vault.v1.MsgActivateProviderResponse")
}
