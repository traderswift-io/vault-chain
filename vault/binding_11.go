package vault

import "github.com/cosmos/gogoproto/proto"

type MsgRegisterFrontend struct {
	Signer           string `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	ProviderId       string `protobuf:"bytes,2,opt,name=provider_id,json=providerId,proto3" json:"provider_id,omitempty"`
	ServicePubkey    []byte `protobuf:"bytes,3,opt,name=service_pubkey,json=servicePubkey,proto3" json:"service_pubkey,omitempty"`
	Endpoint         string `protobuf:"bytes,4,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	ServiceSignature []byte `protobuf:"bytes,5,opt,name=service_signature,json=serviceSignature,proto3" json:"service_signature,omitempty"`
	ServiceSigner    string `protobuf:"bytes,6,opt,name=service_signer,json=serviceSigner,proto3" json:"service_signer,omitempty"`
}

func (message *MsgRegisterFrontend) Reset() {
	*message = MsgRegisterFrontend{}
}

func (message *MsgRegisterFrontend) String() string {
	return proto.CompactTextString(message)
}

func (*MsgRegisterFrontend) ProtoMessage() {}

func (*MsgRegisterFrontend) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{11}
}

func init() {
	proto.RegisterType((*MsgRegisterFrontend)(nil), "traderswift.vault.v1.MsgRegisterFrontend")
}
