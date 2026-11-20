package vault

import "github.com/cosmos/gogoproto/proto"

type MsgSetStorageChoiceResponse struct{}

func (message *MsgSetStorageChoiceResponse) Reset() {
	*message = MsgSetStorageChoiceResponse{}
}

func (message *MsgSetStorageChoiceResponse) String() string {
	return proto.CompactTextString(message)
}

func (*MsgSetStorageChoiceResponse) ProtoMessage() {}

func (*MsgSetStorageChoiceResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{4}
}

func init() {
	proto.RegisterType((*MsgSetStorageChoiceResponse)(nil), "traderswift.vault.v1.MsgSetStorageChoiceResponse")
}
