package entities

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

// normalizeRawValues inspects map values and converts:
// 1. Any []byte (or uint8 slice) of 16 bytes into a valid UUID string.
// 2. Any timestamp strings that lack timezone offsets into RFC3339Nano.
// 3. Any base64-encoded 16-byte binary UUIDs into standard UUID strings.
func normalizeRawValues(val interface{}) interface{} {
	switch v := val.(type) {
	case []byte:
		if len(v) == 16 {
			if u, err := uuid.FromBytes(v); err == nil {
				return u.String()
			}
		}
		return v
	case string:
		// Check for base64 encoded 16-byte UUID (length 24 with ==)
		if len(v) == 24 && strings.HasSuffix(v, "==") {
			if b, err := base64.StdEncoding.DecodeString(v); err == nil && len(b) == 16 {
				if u, err := uuid.FromBytes(b); err == nil {
					return u.String()
				}
			}
		}

		// Check for timestamp missing timezone
		if strings.Contains(v, "T") && len(v) >= 19 && !strings.HasSuffix(v, "Z") && !strings.Contains(v[19:], "+") && !strings.Contains(v[19:], "-") {
			formats := []string{
				"2006-01-02T15:04:05.999999999",
				"2006-01-02T15:04:05",
			}
			for _, f := range formats {
				if t, err := time.ParseInLocation(f, v, time.UTC); err == nil {
					return t.Format(time.RFC3339Nano)
				}
			}
		}
		return v
	case map[string]interface{}:
		for k, item := range v {
			v[k] = normalizeRawValues(item)
		}
		return v
	case []interface{}:
		for i, item := range v {
			v[i] = normalizeRawValues(item)
		}
		return v
	default:
		return v
	}
}

// UnmarshalItem unmarshals a DynamoDB AttributeValue map into a struct via JSON,
// allowing types like uuid.UUID and time.Time that implement encoding.TextUnmarshaler/json.Unmarshaler
// to unmarshal seamlessly without UnmarshalTypeError, even with legacy timestamps or binary UUIDs.
func UnmarshalItem(item map[string]*dynamodb.AttributeValue, target interface{}) error {
	if item == nil {
		return nil
	}
	var raw map[string]interface{}
	if err := dynamodbattribute.UnmarshalMap(item, &raw); err != nil {
		return err
	}
	cleaned := normalizeRawValues(raw)
	data, err := json.Marshal(cleaned)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
