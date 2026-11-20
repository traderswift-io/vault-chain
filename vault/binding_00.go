package vault

import "github.com/cosmos/gogoproto/proto"

type Location struct {
	Name    string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Pointer string `protobuf:"bytes,2,opt,name=pointer,proto3" json:"pointer,omitempty"`
}

func (message *Location) Reset() {
	*message = Location{}
}

func (message *Location) String() string {
	return proto.CompactTextString(message)
}

func (*Location) ProtoMessage() {}

func (*Location) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{0}
}

func init() {
	proto.RegisterType((*Location)(nil), "traderswift.vault.v1.Location")
}
