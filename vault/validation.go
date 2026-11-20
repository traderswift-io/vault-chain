package vault

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"vault-chain/chainprefix"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

const MaxGrantLifetime = uint64(366 * 24 * 60 * 60)

func addressBytes(value, prefix string) ([]byte, error) {
	if len(value) > 90 || value != strings.ToLower(value) {
		return nil, ErrUnauthorized
	}
	actual, data, err := bech32.DecodeAndConvert(value)
	if err != nil || actual != prefix || len(data) != 20 {
		return nil, ErrUnauthorized
	}
	return data, nil
}

func validOwner(owner string) bool {
	_, err := addressBytes(owner, chainprefix.AccountAddressPrefix)
	return err == nil
}

func authorize(signer, owner string) error {
	if signer != owner || !validOwner(owner) {
		return ErrWrongSigner
	}
	return nil
}

func scopeName(value string) bool {
	if len(value) < 1 || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || strings.ContainsRune("._:-", character)) {
			return false
		}
	}
	return true
}

func validDay(day string) bool {
	if len(day) != 10 {
		return false
	}
	parsed, err := time.Parse("2006-01-02", day)
	return err == nil && parsed.Year() >= 1970 && parsed.Format("2006-01-02") == day
}

func validLocations(locations []*Location) bool {
	if len(locations) < 1 || len(locations) > 3 {
		return false
	}
	seen := map[string]bool{}
	for _, location := range locations {
		if location == nil || seen[location.Name] || len(location.Pointer) < 1 || len(location.Pointer) > 256 || !utf8.ValidString(location.Pointer) {
			return false
		}
		if location.Name != "phone" && location.Name != "base" && location.Name != "traderswift" {
			return false
		}
		for _, character := range location.Pointer {
			if character < 32 || character == 127 {
				return false
			}
		}
		seen[location.Name] = true
	}
	return true
}

func validEndpoint(endpoint string) bool {
	if len(endpoint) < 1 || len(endpoint) > 256 || strings.ContainsAny(endpoint, "\x00\r\n\t @?#\\") {
		return false
	}
	parsed, err := url.Parse(endpoint)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" && parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == "" && parsed.RawPath == ""
}

func framedDigest(domain string, fields ...[]byte) []byte {
	encoded := []byte{}
	for _, value := range append([][]byte{[]byte(domain)}, fields...) {
		encoded = binary.BigEndian.AppendUint32(encoded, uint32(len(value)))
		encoded = append(encoded, value...)
	}
	digest := sha256.Sum256(encoded)
	return digest[:]
}

func LinkTranscript(owner, hub string, nonce uint64) []byte {
	return framedDigest(ModuleName+"-link-v2", []byte(chainprefix.ChainID), []byte(owner), []byte(hub), binary.BigEndian.AppendUint64(nil, nonce))
}

func verifyHubLink(message *MsgLinkAccount) bool {
	hub, err := addressBytes(message.HubAccount, chainprefix.HubAddressPrefix)
	if err != nil || len(message.HubSignature) != 64 {
		return false
	}
	var public cryptotypes.PubKey
	switch len(message.HubPublicKey) {
	case 32:
		public = &ed25519.PubKey{Key: message.HubPublicKey}
	case 33:
		public = &secp256k1.PubKey{Key: message.HubPublicKey}
	default:
		return false
	}
	return bytes.Equal(public.Address(), hub) && public.VerifySignature(LinkTranscript(message.Owner, message.HubAccount, message.Nonce), message.HubSignature)
}

func FrontendTranscript(message *MsgRegisterFrontend) []byte {
	body := *message
	body.ServiceSignature = nil
	encoded, _ := json.Marshal(&body)
	return framedDigest(ModuleName+"-frontend-v1", []byte(chainprefix.ChainID), encoded)
}

func GrantCommitment(salt, grantee []byte) []byte {
	if len(salt) != 32 || len(grantee) != 20 {
		return nil
	}
	value := append(append([]byte{}, salt...), grantee...)
	digest := sha256.Sum256(value)
	return digest[:]
}

func blockSeconds(ctx sdk.Context) uint64 {
	if ctx.BlockTime().Unix() < 0 {
		return 0
	}
	return uint64(ctx.BlockTime().Unix())
}
