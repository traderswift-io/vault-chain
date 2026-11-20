package vault

import "github.com/cosmos/gogoproto/proto"

type MsgRemoveProvider struct {
	Signer     string `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	ProviderId string `protobuf:"bytes,2,opt,name=provider_id,json=providerId,proto3" json:"provider_id,omitempty"`
}

func (message *MsgRemoveProvider) Reset() {
	*message = MsgRemoveProvider{}
}

func (message *MsgRemoveProvider) String() string {
	return proto.CompactTextString(message)
}

func (*MsgRemoveProvider) ProtoMessage() {}

func (*MsgRemoveProvider) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{23}
}

func init() {
	proto.RegisterType((*MsgRemoveProvider)(nil), "traderswift.vault.v1.MsgRemoveProvider")
}
