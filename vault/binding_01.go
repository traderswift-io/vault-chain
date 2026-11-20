package vault

import "github.com/cosmos/gogoproto/proto"

type MsgLinkAccount struct {
	Signer       string `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	Owner        string `protobuf:"bytes,2,opt,name=owner,proto3" json:"owner,omitempty"`
	HubAccount   string `protobuf:"bytes,3,opt,name=hub_account,json=hubAccount,proto3" json:"hub_account,omitempty"`
	HubPublicKey []byte `protobuf:"bytes,4,opt,name=hub_public_key,json=hubPublicKey,proto3" json:"hub_public_key,omitempty"`
	HubSignature []byte `protobuf:"bytes,5,opt,name=hub_signature,json=hubSignature,proto3" json:"hub_signature,omitempty"`
	Nonce        uint64 `protobuf:"varint,6,opt,name=nonce,proto3" json:"nonce,omitempty"`
}

func (message *MsgLinkAccount) Reset() {
	*message = MsgLinkAccount{}
}

func (message *MsgLinkAccount) String() string {
	return proto.CompactTextString(message)
}

func (*MsgLinkAccount) ProtoMessage() {}

func (*MsgLinkAccount) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{1}
}

func init() {
	proto.RegisterType((*MsgLinkAccount)(nil), "traderswift.vault.v1.MsgLinkAccount")
}
