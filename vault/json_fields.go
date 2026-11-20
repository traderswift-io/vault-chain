package vault

import (
	"bytes"
	"encoding/json"
	"io"
)

type jsonFrame struct {
	object  bool
	key     bool
	members map[string]bool
	count   int
}

func checkJSONFields(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	frames := []jsonFrame{}
	complete := false
	tokens := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			if complete && len(frames) == 0 {
				return nil
			}
			return ErrBadEncoding
		}
		if err != nil || complete {
			return ErrBadEncoding
		}
		tokens++
		if tokens > MaxRows*32 {
			return ErrLimitExceeded
		}
		if delimiter, ok := token.(json.Delim); ok {
			switch delimiter {
			case '{', '[':
				if len(frames) >= MaxDepth {
					return ErrLimitExceeded
				}
				frames = append(frames, jsonFrame{object: delimiter == '{', key: delimiter == '{', members: map[string]bool{}})
				continue
			case '}', ']':
				if len(frames) == 0 {
					return ErrBadEncoding
				}
				frames = frames[:len(frames)-1]
			}
		} else if len(frames) > 0 && frames[len(frames)-1].object && frames[len(frames)-1].key {
			key, ok := token.(string)
			frame := &frames[len(frames)-1]
			if !ok || len(key) > 128 || frame.members[key] {
				return ErrBadEncoding
			}
			if len(frame.members) >= 128 {
				return ErrLimitExceeded
			}
			frame.members[key], frame.key = true, false
			continue
		}
		if len(frames) == 0 {
			complete = true
		} else {
			frame := &frames[len(frames)-1]
			frame.count++
			if frame.count > MaxRows {
				return ErrLimitExceeded
			}
			if frame.object {
				frame.key = true
			}
		}
	}
}
