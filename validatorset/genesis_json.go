package validatorset

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

func marshalGenesisState(cdc codec.JSONCodec, state *GenesisState) json.RawMessage {
	encoded := cdc.MustMarshalJSON(state)
	if state.PendingAdmin != "" {
		return encoded
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		panic(err)
	}
	delete(fields, "pending_admin")
	encoded, err := json.Marshal(fields)
	if err != nil {
		panic(err)
	}
	return encoded
}
