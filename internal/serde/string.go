package serde

import (
	"encoding/base64"
	"encoding/json"
	"unicode/utf8"
)

type SliceStringSliceString []string

func (s *SliceStringSliceString) MarshalJSON() ([]byte, error) {
	if s == nil {
		return json.Marshal(nil)
	}

	return json.Marshal(*s)
}

func (s *SliceStringSliceString) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var v string
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}

		if v == "" {
			*s = []string{}

			return nil
		}

		*s = []string{v}

		return nil
	}

	var v []string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*s = v

	return nil
}

type StringAsBase64EncodedString string

func (s *StringAsBase64EncodedString) MarshalJSON() ([]byte, error) {
	if s == nil {
		return json.Marshal(nil)
	}

	return json.Marshal(base64.StdEncoding.EncodeToString([]byte(*s)))
}

func (s *StringAsBase64EncodedString) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(v)
	}
	if err != nil || !utf8.Valid(decoded) {
		// Some panels return plain text instead of base64. Preserve it rather
		// than discarding the whole programme guide over one title.
		*s = StringAsBase64EncodedString(v)
		return nil
	}

	*s = StringAsBase64EncodedString(decoded)

	return nil
}
