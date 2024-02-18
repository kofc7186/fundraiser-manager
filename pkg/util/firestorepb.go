package util

import (
	"encoding/json"

	"github.com/googleapis/google-cloudevents-go/cloud/firestoredata"
)

// ParseFirebaseDocument exists because https://github.com/googleapis/google-cloud-go/issues/1438 isn't fixed :/
func ParseFirebaseDocument(from *firestoredata.Document, to interface{}) error {
	if from == nil {
		return nil
	}

	fields := from.GetFields()
	if fields == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range fields {
		result[k] = parseValue(v)
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, to)
}

func parseValue(value *firestoredata.Value) interface{} {
	if value == nil {
		return nil
	}

	switch value.ValueType.(type) {
	case *firestoredata.Value_NullValue:
		return nil
	case *firestoredata.Value_BooleanValue:
		return value.GetBooleanValue()
	case *firestoredata.Value_IntegerValue:
		return value.GetIntegerValue()
	case *firestoredata.Value_DoubleValue:
		return value.GetDoubleValue()
	case *firestoredata.Value_StringValue:
		return value.GetStringValue()
	case *firestoredata.Value_TimestampValue:
		return value.GetTimestampValue().AsTime()
	case *firestoredata.Value_GeoPointValue:
		return value.GetGeoPointValue()
	case *firestoredata.Value_BytesValue:
		return value.GetBytesValue()
	case *firestoredata.Value_ReferenceValue:
		return value.GetReferenceValue()
	case *firestoredata.Value_ArrayValue:
		return parseArray(value.GetArrayValue())
	case *firestoredata.Value_MapValue:
		return parseMap(value.GetMapValue())
	default:
		return nil
	}
}

func parseMap(value *firestoredata.MapValue) interface{} {
	if value == nil {
		return nil
	}

	fields := value.GetFields()
	if fields == nil {
		return nil
	}

	result := make(map[string]interface{})
	for k, v := range fields {
		result[k] = parseValue(v)
	}

	return result
}

func parseArray(value *firestoredata.ArrayValue) interface{} {
	if value == nil {
		return nil
	}

	values := value.GetValues()
	if values == nil {
		return nil
	}

	result := make([]interface{}, len(values))

	for i, v := range values {
		result[i] = parseValue(v)
	}

	return result
}
