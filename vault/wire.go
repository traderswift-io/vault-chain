package vault

import (
	"encoding/binary"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

type WireField struct {
	Number uint64
	Kind   uint64
	Value  uint64
	Bytes  []byte
}

func ScanWire(data []byte, maximum int) ([]WireField, error) {
	if len(data) > maximum {
		return nil, ErrLimitExceeded
	}
	fields := []WireField{}
	var previous uint64
	for len(data) > 0 {
		if len(fields) >= 128 {
			return nil, ErrLimitExceeded
		}
		tag, consumed := binary.Uvarint(data)
		if consumed <= 0 || consumed != len(binary.AppendUvarint(nil, tag)) || tag>>3 == 0 || tag>>3 < previous {
			return nil, ErrBadEncoding
		}
		data = data[consumed:]
		entry := WireField{Number: tag >> 3, Kind: tag & 7}
		previous = entry.Number
		value, size := binary.Uvarint(data)
		if size <= 0 || size != len(binary.AppendUvarint(nil, value)) {
			return nil, ErrBadEncoding
		}
		data = data[size:]
		switch entry.Kind {
		case 0:
			entry.Value = value
		case 2:
			if value > uint64(len(data)) {
				return nil, ErrBadEncoding
			}
			entry.Bytes = data[:int(value)]
			data = data[int(value):]
		default:
			return nil, ErrBadEncoding
		}
		fields = append(fields, entry)
	}
	return fields, nil
}

func ValidateMessageWire(data []byte, message any) error {
	return validateTypedWire(data, reflect.TypeOf(message).Elem(), 1)
}

func validateTypedWire(data []byte, messageType reflect.Type, depth int) error {
	if depth > MaxDepth {
		return ErrLimitExceeded
	}
	fields, err := ScanWire(data, MaxMessageBytes)
	if err != nil {
		return err
	}
	seen := map[uint64]int{}
	for _, wire := range fields {
		if wire.Number > uint64(messageType.NumField()) {
			return ErrBadEncoding
		}
		field := messageType.Field(int(wire.Number) - 1)
		parts := strings.Split(field.Tag.Get("protobuf"), ",")
		if len(parts) < 3 || parts[1] != strconv.FormatUint(wire.Number, 10) {
			return ErrBadEncoding
		}
		repeated := parts[2] == "rep"
		seen[wire.Number]++
		if seen[wire.Number] > 1 && !repeated || seen[wire.Number] > 32 {
			return ErrLimitExceeded
		}
		if field.Name == "SampleResults" && seen[wire.Number] > 4 {
			return ErrLimitExceeded
		}
		if field.Name == "SiblingHashes" && seen[wire.Number] > 24 {
			return ErrLimitExceeded
		}
		target := field.Type
		if repeated {
			target = target.Elem()
		}
		if parts[0] == "varint" {
			if wire.Kind != 0 || wire.Value == 0 && !repeated || target.Kind() == reflect.Bool && wire.Value > 1 {
				return ErrBadEncoding
			}
			continue
		}
		if wire.Kind != 2 || len(wire.Bytes) == 0 && !repeated && target.Kind() != reflect.Pointer {
			return ErrBadEncoding
		}
		if target.Kind() == reflect.Pointer {
			if err := validateTypedWire(wire.Bytes, target.Elem(), depth+1); err != nil {
				return err
			}
		} else if target.Kind() == reflect.String {
			maximum := 90
			if field.Name == "Endpoint" || field.Name == "Pointer" {
				maximum = 256
			}
			if strings.Contains(field.Name, "Micro") || strings.Contains(field.Name, "Rate") || strings.Contains(field.Name, "Numerator") || strings.Contains(field.Name, "Denominator") {
				maximum = 78
			}
			if field.Name == "Pointer" && len(wire.Bytes) > 256 {
				return ErrOversizePointer
			}
			if len(wire.Bytes) > maximum || !utf8.Valid(wire.Bytes) {
				return ErrLimitExceeded
			}
		} else {
			maximum := 33
			if strings.Contains(field.Name, "Signature") {
				maximum = 64
			}
			if strings.Contains(field.Name, "Evidence") || strings.Contains(field.Name, "Authorization") || strings.Contains(field.Name, "Certificate") {
				maximum = 8192
			}
			if len(wire.Bytes) > maximum {
				return ErrLimitExceeded
			}
		}
	}
	return nil
}
