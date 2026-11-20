package vault

import (
	"bytes"
	"encoding/json"
	"unicode/utf8"
)

const (
	MaxMessageBytes     = 32768
	MaxTransactionBytes = 131072
	MaxGenesisBytes     = 16777216
	MaxRows             = 65536
	MaxDepth            = 8
	MaxProviders        = 32
	MaxWork             = 256
)

func CheckJSON(data []byte, maximum int) error {
	if len(data) == 0 || len(data) > maximum || !utf8.Valid(data) {
		return ErrLimitExceeded
	}
	depth, tokens, stringBytes := 0, 0, 0
	quoted, escaped := false, false
	for _, character := range data {
		if quoted {
			stringBytes++
			if stringBytes > MaxMessageBytes {
				return ErrLimitExceeded
			}
			if escaped {
				escaped = false
			} else if character == '\\' {
				escaped = true
			} else if character == '"' {
				quoted = false
			}
			continue
		}
		switch character {
		case '"':
			quoted, stringBytes = true, 0
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		case ',', ':':
			tokens++
		}
		if depth < 0 || depth > MaxDepth || tokens > maximum/2 {
			return ErrLimitExceeded
		}
	}
	if quoted || depth != 0 {
		return ErrBadEncoding
	}
	return checkJSONFields(data)
}

func DecodeCanonical(data []byte, target any, maximum int) error {
	if err := CheckJSON(data, maximum); err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return ErrBadEncoding
	}
	canonical, err := json.Marshal(target)
	if err != nil || !bytes.Equal(canonical, data) {
		return ErrBadEncoding
	}
	return nil
}

func identifier(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '-' || character == '_') {
			return false
		}
	}
	return true
}

func hashValue(value []byte) bool { return len(value) == 32 }

func safeAdd(left, right uint64) (uint64, error) {
	if ^uint64(0)-left < right {
		return 0, ErrLimitExceeded
	}
	return left + right, nil
}
