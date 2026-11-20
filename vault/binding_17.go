package vault

import "github.com/cosmos/gogoproto/proto"

type MsgActivateProvider struct {
	Signer             string `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	ProviderId         string `protobuf:"bytes,2,opt,name=provider_id,json=providerId,proto3" json:"provider_id,omitempty"`
	ConsensusPubkey    []byte `protobuf:"bytes,3,opt,name=consensus_pubkey,json=consensusPubkey,proto3" json:"consensus_pubkey,omitempty"`
	ConsensusSignature []byte `protobuf:"bytes,4,opt,name=consensus_signature,json=consensusSignature,proto3" json:"consensus_signature,omitempty"`
}

func (message *MsgActivateProvider) Reset() {
	*message = MsgActivateProvider{}
}

func (message *MsgActivateProvider) String() string {
	return proto.CompactTextString(message)
}

func (*MsgActivateProvider) ProtoMessage() {}

func (*MsgActivateProvider) Descriptor() ([]byte, []int) {
	return moduleDescriptor, []int{17}
}

func init() {
	proto.RegisterType((*MsgActivateProvider)(nil), "traderswift.vault.v1.MsgActivateProvider")
}
