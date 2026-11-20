package vault

import "github.com/cosmos/gogoproto/proto"

type MsgCommitRoot struct {
	Signer string `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	Owner  string `protobuf:"bytes,2,opt,name=owner,proto3" json:"owner,omitempty"`
	Day    string `protobuf:"bytes,3,opt,name=day,proto3" json:"day,omitempty"`
	Root   []byte `protobuf:"bytes,4,opt,name=root,proto3" json:"root,omitempty"`
}

func (message *MsgCommitRoot) Reset() {
	*message = MsgCommitRoot{}
}

func (message *MsgCommitRoot) String() string {
	return proto.CompactTextString(message)
}

func (*MsgCommitRoot) ProtoMessage() {}

func (*MsgCommitRoot) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{5}
}

func init() {
	proto.RegisterType((*MsgCommitRoot)(nil), "traderswift.vault.v1.MsgCommitRoot")
}
