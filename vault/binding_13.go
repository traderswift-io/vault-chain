package vault

import "github.com/cosmos/gogoproto/proto"

type MsgUpdateParams struct {
	Signer             string `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	BaseEnabled        bool   `protobuf:"varint,2,opt,name=base_enabled,json=baseEnabled,proto3" json:"base_enabled,omitempty"`
	TraderswiftEnabled bool   `protobuf:"varint,3,opt,name=traderswift_enabled,json=traderswiftEnabled,proto3" json:"traderswift_enabled,omitempty"`
}

func (message *MsgUpdateParams) Reset() {
	*message = MsgUpdateParams{}
}

func (message *MsgUpdateParams) String() string {
	return proto.CompactTextString(message)
}

func (*MsgUpdateParams) ProtoMessage() {}

func (*MsgUpdateParams) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{13}
}

func init() {
	proto.RegisterType((*MsgUpdateParams)(nil), "traderswift.vault.v1.MsgUpdateParams")
}
