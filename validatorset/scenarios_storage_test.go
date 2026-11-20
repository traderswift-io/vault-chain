package validatorset

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func syntheticValidators() []Validator {
	result := []Validator{}
	for index := 0; index < 4; index++ {
		seed := sha256.Sum256([]byte(fmt.Sprintf("synthetic-consensus-fixture-%d", index)))
		public := ed25519.NewKeyFromSeed(seed[:]).Public().(ed25519.PublicKey)
		result = append(result, Validator{ConsensusPublicKey: base64.StdEncoding.EncodeToString(public), Moniker: fmt.Sprintf("founder-%d", index), Power: 100, Founding: true})
	}
	return result
}

func TestAggregateMembershipCap(t *testing.T) {
	founders := syntheticValidators()
	if err := requireSetWithinOutsiderLimit(founders); err != nil {
		t.Fatal(err)
	}
	outsider := founders[0]
	outsider.Founding, outsider.Power = false, 200
	if requireSetWithinOutsiderLimit(append(founders, outsider)) == nil {
		t.Fatal("one-third accepted")
	}
	outsider.Power = 50
	if requireSetWithinOutsiderLimit(append(founders, outsider)) == nil {
		t.Fatal("founder-loss boundary accepted")
	}
	outsider.Power = 49
	if err := requireSetWithinOutsiderLimit(append(founders, outsider)); err != nil {
		t.Fatal(err)
	}
	outsider.Power = 25
	if requireSetWithinOutsiderLimit(append(founders, outsider, outsider)) == nil {
		t.Fatal("aggregate outsider cap bypassed")
	}
	server := NewMsgServer(nil)
	if _, err := server.AddValidator(nil, &MsgAddValidator{Founding: true}); err != ErrUnauthorized {
		t.Fatal("administrative add bypass")
	}
	if _, err := server.SetValidatorPower(nil, &MsgSetValidatorPower{Power: 1}); err != ErrUnauthorized {
		t.Fatal("administrative reweight bypass")
	}
	if _, err := server.RemoveValidator(nil, &MsgRemoveValidator{}); err != ErrUnauthorized {
		t.Fatal("administrative removal bypass")
	}
	fmt.Println("PASS validatorset aggregate cap and administrative bypass refusals")
}

func TestBoundedValidatorParsers(t *testing.T) {
	if _, err := decodeConsensusPublicKey(strings.Repeat("A", 1000)); err == nil {
		t.Fatal("oversize base64 accepted")
	}
	random := rand.New(rand.NewSource(18))
	for index := 0; index < 256; index++ {
		data := make([]byte, random.Intn(128))
		_, _ = random.Read(data)
		_, _ = decodeConsensusPublicKey(string(data))
		var validator Validator
		_ = validator.Unmarshal(data)
		var genesis GenesisState
		_ = genesis.Unmarshal(data)
	}
	fmt.Println("PASS validatorset bounded parser fuzz")
}
