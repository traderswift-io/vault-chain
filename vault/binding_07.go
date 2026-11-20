package vault

import "github.com/cosmos/gogoproto/proto"

type MsgGrant struct {
	Signer            string `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	Owner             string `protobuf:"bytes,2,opt,name=owner,proto3" json:"owner,omitempty"`
	GranteeCommitment []byte `protobuf:"bytes,3,opt,name=grantee_commitment,json=granteeCommitment,proto3" json:"grantee_commitment,omitempty"`
	Scope             string `protobuf:"bytes,4,opt,name=scope,proto3" json:"scope,omitempty"`
	Expiry            uint64 `protobuf:"varint,5,opt,name=expiry,proto3" json:"expiry,omitempty"`
}

func (message *MsgGrant) Reset() {
	*message = MsgGrant{}
}

func (message *MsgGrant) String() string {
	return proto.CompactTextString(message)
}

func (*MsgGrant) ProtoMessage() {}

func (*MsgGrant) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{7}
}

func init() {
	proto.RegisterType((*MsgGrant)(nil), "traderswift.vault.v1.MsgGrant")
}
