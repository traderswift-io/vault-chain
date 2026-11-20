package validatorset

import (
	"math/rand"
	"testing"
)

func TestEveryInheritedDecoder(t *testing.T) {
	decoders := []func([]byte) error{
		(&GenesisState{}).Unmarshal,
		(&Validator{}).Unmarshal,
		(&MsgSetValidatorPower{}).Unmarshal,
		(&MsgSetValidatorPowerResponse{}).Unmarshal,
		(&MsgRemoveValidator{}).Unmarshal,
		(&MsgRemoveValidatorResponse{}).Unmarshal,
		(&QueryAdminRequest{}).Unmarshal,
		(&QueryAdminResponse{}).Unmarshal,
		(&QueryValidatorsRequest{}).Unmarshal,
		(&MsgSetAdminResponse{}).Unmarshal,
		(&MsgAcceptAdmin{}).Unmarshal,
		(&MsgAcceptAdminResponse{}).Unmarshal,
		(&MsgAddValidator{}).Unmarshal,
		(&MsgAddValidatorResponse{}).Unmarshal,
		(&MsgSetAdmin{}).Unmarshal,
		(&QueryValidatorsResponse{}).Unmarshal,
		(&QueryValidatorRequest{}).Unmarshal,
		(&QueryValidatorResponse{}).Unmarshal,
	}
	random := rand.New(rand.NewSource(1819))
	for _, decoder := range decoders {
		for iteration := 0; iteration < 128; iteration++ {
			data := make([]byte, random.Intn(128))
			_, _ = random.Read(data)
			_ = decoder(data)
		}
	}
}
