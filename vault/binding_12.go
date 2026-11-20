package vault

import "github.com/cosmos/gogoproto/proto"

type MsgRegisterFrontendResponse struct{}

func (message *MsgRegisterFrontendResponse) Reset() {
	*message = MsgRegisterFrontendResponse{}
}

func (message *MsgRegisterFrontendResponse) String() string {
	return proto.CompactTextString(message)
}

func (*MsgRegisterFrontendResponse) ProtoMessage() {}

func (*MsgRegisterFrontendResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{12}
}

func init() {
	proto.RegisterType((*MsgRegisterFrontendResponse)(nil), "traderswift.vault.v1.MsgRegisterFrontendResponse")
}
