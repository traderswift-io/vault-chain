package vault

import "github.com/cosmos/gogoproto/proto"

type MsgCommitRootResponse struct{}

func (message *MsgCommitRootResponse) Reset() {
	*message = MsgCommitRootResponse{}
}

func (message *MsgCommitRootResponse) String() string {
	return proto.CompactTextString(message)
}

func (*MsgCommitRootResponse) ProtoMessage() {}

func (*MsgCommitRootResponse) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{6}
}

func init() {
	proto.RegisterType((*MsgCommitRootResponse)(nil), "traderswift.vault.v1.MsgCommitRootResponse")
}
