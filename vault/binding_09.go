package vault

import "github.com/cosmos/gogoproto/proto"

type MsgRevoke struct {
	Signer            string `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	Owner             string `protobuf:"bytes,2,opt,name=owner,proto3" json:"owner,omitempty"`
	GranteeCommitment []byte `protobuf:"bytes,3,opt,name=grantee_commitment,json=granteeCommitment,proto3" json:"grantee_commitment,omitempty"`
	Scope             string `protobuf:"bytes,4,opt,name=scope,proto3" json:"scope,omitempty"`
}

func (message *MsgRevoke) Reset() {
	*message = MsgRevoke{}
}

func (message *MsgRevoke) String() string {
	return proto.CompactTextString(message)
}

func (*MsgRevoke) ProtoMessage() {}

func (*MsgRevoke) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{9}
}

func init() {
	proto.RegisterType((*MsgRevoke)(nil), "traderswift.vault.v1.MsgRevoke")
}
