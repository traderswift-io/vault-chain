package vault

import "github.com/cosmos/gogoproto/proto"

type MsgReplaceProviderOwner struct {
	Signer     string `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	ProviderId string `protobuf:"bytes,2,opt,name=provider_id,json=providerId,proto3" json:"provider_id,omitempty"`
	NewOwner   string `protobuf:"bytes,3,opt,name=new_owner,json=newOwner,proto3" json:"new_owner,omitempty"`
}

func (message *MsgReplaceProviderOwner) Reset() {
	*message = MsgReplaceProviderOwner{}
}

func (message *MsgReplaceProviderOwner) String() string {
	return proto.CompactTextString(message)
}

func (*MsgReplaceProviderOwner) ProtoMessage() {}

func (*MsgReplaceProviderOwner) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{25}
}

func init() {
	proto.RegisterType((*MsgReplaceProviderOwner)(nil), "traderswift.vault.v1.MsgReplaceProviderOwner")
}
