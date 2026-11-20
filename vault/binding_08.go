package vault

import "github.com/cosmos/gogoproto/proto"

type MsgGrantResponse struct{}

func (message *MsgGrantResponse) Reset() {
	*message = MsgGrantResponse{}
}

func (message *MsgGrantResponse) String() string {
	return proto.CompactTextString(message)
}

func (*MsgGrantResponse) ProtoMessage() {}

func (*MsgGrantResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{8}
}

func init() {
	proto.RegisterType((*MsgGrantResponse)(nil), "traderswift.vault.v1.MsgGrantResponse")
}
