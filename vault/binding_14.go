package vault

import "github.com/cosmos/gogoproto/proto"

type MsgUpdateParamsResponse struct{}

func (message *MsgUpdateParamsResponse) Reset() {
	*message = MsgUpdateParamsResponse{}
}

func (message *MsgUpdateParamsResponse) String() string {
	return proto.CompactTextString(message)
}

func (*MsgUpdateParamsResponse) ProtoMessage() {}

func (*MsgUpdateParamsResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{14}
}

func init() {
	proto.RegisterType((*MsgUpdateParamsResponse)(nil), "traderswift.vault.v1.MsgUpdateParamsResponse")
}
