package chainprefix

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var applyOnce sync.Once

func Apply() {
	applyOnce.Do(func() {
		config := sdk.GetConfig()
		config.SetBech32PrefixForAccount(AccountAddressPrefix, AccountAddressPrefix+sdk.PrefixPublic)
		config.SetBech32PrefixForValidator(AccountAddressPrefix+sdk.PrefixValidator+sdk.PrefixOperator, AccountAddressPrefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic)
		config.SetBech32PrefixForConsensusNode(AccountAddressPrefix+sdk.PrefixValidator+sdk.PrefixConsensus, AccountAddressPrefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic)
		config.Seal()
	})
}
