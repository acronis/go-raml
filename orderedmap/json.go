package orderedmap

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"unicode/utf8"

	"strconv"

	"github.com/buger/jsonparser"
)

var (
	_ json.Marshaler   = &OrderedMap[int, any]{}
	_ json.Unmarshaler = &OrderedMap[int, any]{}
)

// MarshalJSON implements the json.Marshaler interface.
func (om *OrderedMap[K, V]) MarshalJSON() ([]byte, error) { //nolint:funlen
	if om == nil || om.list == nil {
		return []byte("null"), nil
	}

	var buf bytes.Buffer
	buf.WriteByte('{')

	for pair, firstIteration := om.Oldest(), true; pair != nil; pair = pair.Next() {
		if firstIteration {
			firstIteration = false
		} else {
			buf.WriteByte(',')
		}

		switch key := any(pair.Key).(type) {
		case string:
			writeJSONString(&buf, key)
		case encoding.TextMarshaler:
			text, err := key.MarshalText()
			if err != nil {
				return nil, err
			}
			buf.WriteByte('"')
			buf.Write(text)
			buf.WriteByte('"')
		case int:
			writeQuotedInt(&buf, int64(key))
		case int8:
			writeQuotedInt(&buf, int64(key))
		case int16:
			writeQuotedInt(&buf, int64(key))
		case int32:
			writeQuotedInt(&buf, int64(key))
		case int64:
			writeQuotedInt(&buf, key)
		case uint:
			writeQuotedUint(&buf, uint64(key))
		case uint8:
			writeQuotedUint(&buf, uint64(key))
		case uint16:
			writeQuotedUint(&buf, uint64(key))
		case uint32:
			writeQuotedUint(&buf, uint64(key))
		case uint64:
			writeQuotedUint(&buf, key)
		default:

			// this switch takes care of wrapper types around primitive types, such as
			// type myType string
			switch keyValue := reflect.ValueOf(key); keyValue.Type().Kind() {
			case reflect.String:
				writeJSONString(&buf, keyValue.String())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				writeQuotedInt(&buf, keyValue.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				writeQuotedUint(&buf, keyValue.Uint())
			default:
				return nil, fmt.Errorf("unsupported key type: %T", key)
			}
		}

		buf.WriteByte(':')
		value, err := json.Marshal(pair.Value)
		if err != nil {
			return nil, err
		}
		buf.Write(value)
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

func writeQuotedInt(buf *bytes.Buffer, n int64) {
	buf.WriteByte('"')
	buf.WriteString(strconv.FormatInt(n, 10))
	buf.WriteByte('"')
}

func writeQuotedUint(buf *bytes.Buffer, n uint64) {
	buf.WriteByte('"')
	buf.WriteString(strconv.FormatUint(n, 10))
	buf.WriteByte('"')
}

const hexChars = "0123456789abcdef"

// writeJSONString writes s as a JSON string with HTML-safe escaping.
// The output matches github.com/mailru/easyjson/jwriter.Writer.String used by github.com/wk8/go-ordered-map/v2:
// \t, \r, \n, \\ and \" use short escapes, other control characters and <, >, & use \u00XX,
// invalid UTF-8 bytes become \ufffd, U+2028 and U+2029 are escaped.
func writeJSONString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')

	p := 0 // start of the pending part that needs no escaping
	for i := 0; i < len(s); {
		c := s[i]

		if c < utf8.RuneSelf {
			if c >= 0x20 && c != '"' && c != '\\' && c != '<' && c != '>' && c != '&' {
				i++
				continue
			}

			buf.WriteString(s[p:i])
			switch c {
			case '\t':
				buf.WriteString(`\t`)
			case '\r':
				buf.WriteString(`\r`)
			case '\n':
				buf.WriteString(`\n`)
			case '\\':
				buf.WriteString(`\\`)
			case '"':
				buf.WriteString(`\"`)
			default:
				buf.WriteString(`\u00`)
				buf.WriteByte(hexChars[c>>4])
				buf.WriteByte(hexChars[c&0xf])
			}

			i++
			p = i
			continue
		}

		r, width := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && width == 1 {
			buf.WriteString(s[p:i])
			buf.WriteString(`\ufffd`)
			i++
			p = i
			continue
		}

		if r == '\u2028' || r == '\u2029' {
			buf.WriteString(s[p:i])
			buf.WriteString(`\u202`)
			buf.WriteByte(hexChars[r&0xf])
			i += width
			p = i
			continue
		}
		i += width
	}
	buf.WriteString(s[p:])
	buf.WriteByte('"')
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (om *OrderedMap[K, V]) UnmarshalJSON(data []byte) error {
	if om.list == nil {
		om.initialize(0)
	}

	return jsonparser.ObjectEach(
		data,
		func(keyData []byte, valueData []byte, dataType jsonparser.ValueType, offset int) error {
			if dataType == jsonparser.String {
				// jsonparser removes the enclosing quotes; we need to restore them to make a valid JSON
				valueData = data[offset-len(valueData)-2 : offset]
			}

			var key K
			var value V

			switch typedKey := any(&key).(type) {
			case *string:
				s, err := decodeUTF8(keyData)
				if err != nil {
					return err
				}
				*typedKey = s
			case encoding.TextUnmarshaler:
				if err := typedKey.UnmarshalText(keyData); err != nil {
					return err
				}
			case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
				if err := json.Unmarshal(keyData, typedKey); err != nil {
					return err
				}
			default:
				// this switch takes care of wrapper types around primitive types, such as
				// type myType string
				switch reflect.TypeOf(key).Kind() {
				case reflect.String:
					s, err := decodeUTF8(keyData)
					if err != nil {
						return err
					}

					convertedKeyData := reflect.ValueOf(s).Convert(reflect.TypeOf(key))
					reflect.ValueOf(&key).Elem().Set(convertedKeyData)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if err := json.Unmarshal(keyData, &key); err != nil {
						return err
					}
				default:
					return fmt.Errorf("unsupported key type: %T", key)
				}
			}

			if err := json.Unmarshal(valueData, &value); err != nil {
				return err
			}

			om.Set(key, value)
			return nil
		})
}

func decodeUTF8(input []byte) (string, error) {
	remaining, offset := input, 0
	runes := make([]rune, 0, len(remaining))

	for len(remaining) > 0 {
		r, size := utf8.DecodeRune(remaining)
		if r == utf8.RuneError && size <= 1 {
			return "", fmt.Errorf("not a valid UTF-8 string (at position %d): %s", offset, string(input))
		}

		runes = append(runes, r)
		remaining = remaining[size:]
		offset += size
	}

	return string(runes), nil
}
