// Copyright Dit 2026
// SPDX-License-Identifier: BUSL-1.1

// Package conversion provides protobuf struct conversion utilities.
package conversion

import (
	"fmt"
	"reflect"

	"github.com/fatih/structs"
	protobuf_struct "github.com/golang/protobuf/ptypes/struct"
)

// decodeValue converts a single protobuf Value into its Go representation:
// NullValue -> nil, NumberValue -> float64, StringValue -> string,
// BoolValue -> bool, StructValue -> map[string]interface{}, and
// ListValue -> []interface{}. A nil Value decodes to nil; a Value whose
// Kind is unset is an error.
func decodeValue(value *protobuf_struct.Value) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch kind := value.GetKind().(type) {
	case *protobuf_struct.Value_NullValue:
		return nil, nil
	case *protobuf_struct.Value_NumberValue:
		return kind.NumberValue, nil
	case *protobuf_struct.Value_StringValue:
		return kind.StringValue, nil
	case *protobuf_struct.Value_BoolValue:
		return kind.BoolValue, nil
	case *protobuf_struct.Value_StructValue:
		return decodeFields(kind.StructValue.GetFields())
	case *protobuf_struct.Value_ListValue:
		values := kind.ListValue.GetValues()
		result := make([]interface{}, len(values))
		for i, element := range values {
			decoded, err := decodeValue(element)
			if err != nil {
				return nil, err
			}
			result[i] = decoded
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert the value %+v", value)
	}
}

// decodeFields converts a protobuf Struct field map into a Go map.
func decodeFields(fields map[string]*protobuf_struct.Value) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(fields))
	for key, value := range fields {
		decoded, err := decodeValue(value)
		if err != nil {
			return nil, err
		}
		result[key] = decoded
	}
	return result, nil
}

// Struct2Map converts a protobuf Struct to a Go map[string]interface{}.
// A nil Struct converts to an empty map.
func Struct2Map(str *protobuf_struct.Struct) (map[string]interface{}, error) {
	return decodeFields(str.GetFields())
}

// encodeValue converts a Go value into a protobuf Value. Exact string and
// bool types are matched first; remaining values are handled by reflection
// so that numeric types of any width (including named types) become
// NumberValue. Named string and bool types are deliberately rejected: the
// wire representation would silently discard the type, so callers must
// convert explicitly.
func encodeValue(entry interface{}) (*protobuf_struct.Value, error) {
	switch typed := entry.(type) {
	case nil:
		return &protobuf_struct.Value{Kind: &protobuf_struct.Value_NullValue{}}, nil
	case string:
		return &protobuf_struct.Value{Kind: &protobuf_struct.Value_StringValue{StringValue: typed}}, nil
	case bool:
		return &protobuf_struct.Value{Kind: &protobuf_struct.Value_BoolValue{BoolValue: typed}}, nil
	}

	reflected := reflect.ValueOf(entry)
	switch reflected.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return numberValue(float64(reflected.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return numberValue(float64(reflected.Uint())), nil
	case reflect.Float32, reflect.Float64:
		return numberValue(reflected.Float()), nil
	case reflect.String:
		return nil, fmt.Errorf("cannot convert string value")
	case reflect.Bool:
		return nil, fmt.Errorf("cannot convert boolean value")
	case reflect.Array, reflect.Slice:
		values := make([]*protobuf_struct.Value, reflected.Len())
		for i := range values {
			encoded, err := encodeValue(reflected.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			values[i] = encoded
		}
		return &protobuf_struct.Value{Kind: &protobuf_struct.Value_ListValue{
			ListValue: &protobuf_struct.ListValue{Values: values},
		}}, nil
	case reflect.Map:
		asMap := make(map[string]interface{}, reflected.Len())
		for _, key := range reflected.MapKeys() {
			asMap[key.String()] = reflected.MapIndex(key).Interface()
		}
		encoded, err := Map2Struct(asMap)
		if err != nil {
			return nil, err
		}
		return &protobuf_struct.Value{Kind: &protobuf_struct.Value_StructValue{StructValue: encoded}}, nil
	case reflect.Struct:
		return encodeValue(structs.Map(entry))
	default:
		return nil, fmt.Errorf("cannot convert [%+v] kind:%s", entry, reflected.Kind())
	}
}

func numberValue(number float64) *protobuf_struct.Value {
	return &protobuf_struct.Value{Kind: &protobuf_struct.Value_NumberValue{NumberValue: number}}
}

// Map2Struct converts a Go map[string]interface{} to a protobuf Struct.
func Map2Struct(input map[string]interface{}) (*protobuf_struct.Struct, error) {
	result := &protobuf_struct.Struct{Fields: make(map[string]*protobuf_struct.Value, len(input))}
	for key, value := range input {
		encoded, err := encodeValue(value)
		if err != nil {
			return nil, err
		}
		result.Fields[key] = encoded
	}
	return result, nil
}

// Struct2ProtobufStruct converts any interface{} to a protobuf Struct.
func Struct2ProtobufStruct(input interface{}) (*protobuf_struct.Struct, error) {
	return Map2Struct(structs.Map(input))
}
