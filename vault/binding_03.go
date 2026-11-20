package vault

import "github.com/cosmos/gogoproto/proto"

type MsgSetStorageChoice struct {
	Signer    string      `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	Owner     string      `protobuf:"bytes,2,opt,name=owner,proto3" json:"owner,omitempty"`
	Locations []*Location `protobuf:"bytes,3,rep,name=locations,proto3" json:"locations,omitempty"`
}

func (message *MsgSetStorageChoice) Reset() {
	*message = MsgSetStorageChoice{}
}

func (message *MsgSetStorageChoice) String() string {
	return proto.CompactTextString(message)
}

func (*MsgSetStorageChoice) ProtoMessage() {}

func (*MsgSetStorageChoice) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{3}
}

func init() {
	proto.RegisterType((*MsgSetStorageChoice)(nil), "traderswift.vault.v1.MsgSetStorageChoice")
}
