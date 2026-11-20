package vault

import "github.com/cosmos/gogoproto/proto"

type MsgRemoveFrontend struct {
	Signer     string `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	ProviderId string `protobuf:"bytes,2,opt,name=provider_id,json=providerId,proto3" json:"provider_id,omitempty"`
}

func (message *MsgRemoveFrontend) Reset() {
	*message = MsgRemoveFrontend{}
}

func (message *MsgRemoveFrontend) String() string {
	return proto.CompactTextString(message)
}

func (*MsgRemoveFrontend) ProtoMessage() {}

func (*MsgRemoveFrontend) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{15}
}

func init() {
	proto.RegisterType((*MsgRemoveFrontend)(nil), "traderswift.vault.v1.MsgRemoveFrontend")
}
