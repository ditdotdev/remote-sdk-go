// Package conversion tests cover the bidirectional protobuf Struct
// <-> Go map conversion utilities.
package conversion

import (
	"testing"

	pb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
)

// === Map2Struct: Go map -> protobuf Struct ===

func TestMap2Struct_Empty(t *testing.T) {
	s, err := Map2Struct(map[string]interface{}{})
	if assert.NoError(t, err) {
		assert.NotNil(t, s)
		assert.Empty(t, s.Fields)
	}
}

func TestMap2Struct_String(t *testing.T) {
	s, err := Map2Struct(map[string]interface{}{"k": "hello"})
	if assert.NoError(t, err) {
		assert.Equal(t, "hello", s.Fields["k"].GetStringValue())
	}
}

func TestMap2Struct_AllIntWidths(t *testing.T) {
	s, err := Map2Struct(map[string]interface{}{
		"int":   42,
		"int8":  int8(8),
		"int16": int16(16),
		"int32": int32(32),
		"int64": int64(64),
	})
	if assert.NoError(t, err) {
		assert.Equal(t, float64(42), s.Fields["int"].GetNumberValue())
		assert.Equal(t, float64(8), s.Fields["int8"].GetNumberValue())
		assert.Equal(t, float64(16), s.Fields["int16"].GetNumberValue())
		assert.Equal(t, float64(32), s.Fields["int32"].GetNumberValue())
		assert.Equal(t, float64(64), s.Fields["int64"].GetNumberValue())
	}
}

func TestMap2Struct_AllUintWidths(t *testing.T) {
	s, err := Map2Struct(map[string]interface{}{
		"uint":   uint(1),
		"uint8":  uint8(2),
		"uint16": uint16(3),
		"uint32": uint32(4),
		"uint64": uint64(5),
	})
	if assert.NoError(t, err) {
		assert.Equal(t, float64(1), s.Fields["uint"].GetNumberValue())
		assert.Equal(t, float64(2), s.Fields["uint8"].GetNumberValue())
		assert.Equal(t, float64(3), s.Fields["uint16"].GetNumberValue())
		assert.Equal(t, float64(4), s.Fields["uint32"].GetNumberValue())
		assert.Equal(t, float64(5), s.Fields["uint64"].GetNumberValue())
	}
}

func TestMap2Struct_Floats(t *testing.T) {
	s, err := Map2Struct(map[string]interface{}{
		"f32": float32(1.5),
		"f64": 2.5,
	})
	if assert.NoError(t, err) {
		assert.InDelta(t, 1.5, s.Fields["f32"].GetNumberValue(), 0.0001)
		assert.Equal(t, 2.5, s.Fields["f64"].GetNumberValue())
	}
}

func TestMap2Struct_Bool(t *testing.T) {
	s, err := Map2Struct(map[string]interface{}{"yes": true, "no": false})
	if assert.NoError(t, err) {
		assert.True(t, s.Fields["yes"].GetBoolValue())
		assert.False(t, s.Fields["no"].GetBoolValue())
	}
}

func TestMap2Struct_Nil(t *testing.T) {
	s, err := Map2Struct(map[string]interface{}{"none": nil})
	if assert.NoError(t, err) {
		assert.IsType(t, &pb.Value_NullValue{}, s.Fields["none"].GetKind())
	}
}

func TestMap2Struct_Slice(t *testing.T) {
	s, err := Map2Struct(map[string]interface{}{"items": []string{"a", "b", "c"}})
	if assert.NoError(t, err) {
		list := s.Fields["items"].GetListValue()
		if assert.NotNil(t, list) && assert.Len(t, list.Values, 3) {
			assert.Equal(t, "a", list.Values[0].GetStringValue())
			assert.Equal(t, "b", list.Values[1].GetStringValue())
			assert.Equal(t, "c", list.Values[2].GetStringValue())
		}
	}
}

func TestMap2Struct_Array(t *testing.T) {
	arr := [2]int{10, 20}
	s, err := Map2Struct(map[string]interface{}{"items": arr})
	if assert.NoError(t, err) {
		list := s.Fields["items"].GetListValue()
		if assert.NotNil(t, list) && assert.Len(t, list.Values, 2) {
			assert.Equal(t, float64(10), list.Values[0].GetNumberValue())
			assert.Equal(t, float64(20), list.Values[1].GetNumberValue())
		}
	}
}

func TestMap2Struct_NestedMap(t *testing.T) {
	s, err := Map2Struct(map[string]interface{}{
		"outer": map[string]interface{}{"inner": "value"},
	})
	if assert.NoError(t, err) {
		outer := s.Fields["outer"].GetStructValue()
		if assert.NotNil(t, outer) {
			assert.Equal(t, "value", outer.Fields["inner"].GetStringValue())
		}
	}
}

func TestMap2Struct_UnsupportedType(t *testing.T) {
	// channels are not a supported reflect.Kind in elabEntry
	s, err := Map2Struct(map[string]interface{}{"ch": make(chan int)})
	assert.Error(t, err)
	assert.Nil(t, s)
}

func TestMap2Struct_TypeAliasedString(t *testing.T) {
	// reflect.Kind reports String, but the .(string) assertion fails.
	type customStr string
	_, err := Map2Struct(map[string]interface{}{"k": customStr("hi")})
	assert.Error(t, err)
}

func TestMap2Struct_TypeAliasedBool(t *testing.T) {
	// reflect.Kind reports Bool, but the .(bool) assertion fails.
	type customBool bool
	_, err := Map2Struct(map[string]interface{}{"k": customBool(true)})
	assert.Error(t, err)
}

func TestMap2Struct_ErrorBubblesFromList(t *testing.T) {
	_, err := Map2Struct(map[string]interface{}{
		"list": []interface{}{make(chan int)},
	})
	assert.Error(t, err)
}

func TestMap2Struct_ErrorBubblesFromNestedMap(t *testing.T) {
	_, err := Map2Struct(map[string]interface{}{
		"outer": map[string]interface{}{"bad": make(chan int)},
	})
	assert.Error(t, err)
}

// === Struct2Map: protobuf Struct -> Go map ===

func TestStruct2Map_NilInput(t *testing.T) {
	m, err := Struct2Map(nil)
	if assert.NoError(t, err) {
		assert.Empty(t, m)
	}
}

func TestStruct2Map_Empty(t *testing.T) {
	m, err := Struct2Map(&pb.Struct{Fields: map[string]*pb.Value{}})
	if assert.NoError(t, err) {
		assert.Empty(t, m)
	}
}

func TestStruct2Map_Scalars(t *testing.T) {
	s := &pb.Struct{Fields: map[string]*pb.Value{
		"str":  {Kind: &pb.Value_StringValue{StringValue: "hello"}},
		"num":  {Kind: &pb.Value_NumberValue{NumberValue: 42.5}},
		"yes":  {Kind: &pb.Value_BoolValue{BoolValue: true}},
		"no":   {Kind: &pb.Value_BoolValue{BoolValue: false}},
		"null": {Kind: &pb.Value_NullValue{}},
	}}
	m, err := Struct2Map(s)
	if assert.NoError(t, err) {
		assert.Equal(t, "hello", m["str"])
		assert.Equal(t, 42.5, m["num"])
		assert.Equal(t, true, m["yes"])
		assert.Equal(t, false, m["no"])
		assert.Nil(t, m["null"])
	}
}

func TestStruct2Map_NilFieldValue(t *testing.T) {
	// elabValue returns (nil, nil) when given a nil *Value
	s := &pb.Struct{Fields: map[string]*pb.Value{"k": nil}}
	m, err := Struct2Map(s)
	if assert.NoError(t, err) {
		assert.Nil(t, m["k"])
	}
}

func TestStruct2Map_NestedStruct(t *testing.T) {
	nested := &pb.Struct{Fields: map[string]*pb.Value{
		"inner": {Kind: &pb.Value_StringValue{StringValue: "v"}},
	}}
	s := &pb.Struct{Fields: map[string]*pb.Value{
		"outer": {Kind: &pb.Value_StructValue{StructValue: nested}},
	}}
	m, err := Struct2Map(s)
	if assert.NoError(t, err) {
		outer, ok := m["outer"].(map[string]interface{})
		if assert.True(t, ok) {
			assert.Equal(t, "v", outer["inner"])
		}
	}
}

func TestStruct2Map_List(t *testing.T) {
	list := &pb.ListValue{Values: []*pb.Value{
		{Kind: &pb.Value_StringValue{StringValue: "x"}},
		{Kind: &pb.Value_NumberValue{NumberValue: 1}},
		{Kind: &pb.Value_BoolValue{BoolValue: true}},
	}}
	s := &pb.Struct{Fields: map[string]*pb.Value{
		"items": {Kind: &pb.Value_ListValue{ListValue: list}},
	}}
	m, err := Struct2Map(s)
	if assert.NoError(t, err) {
		items, ok := m["items"].([]interface{})
		if assert.True(t, ok) && assert.Len(t, items, 3) {
			assert.Equal(t, "x", items[0])
			assert.Equal(t, float64(1), items[1])
			assert.Equal(t, true, items[2])
		}
	}
}

// === Struct2ProtobufStruct: arbitrary Go struct -> protobuf Struct ===

type sampleStruct struct {
	Name string
	Age  int
}

func TestStruct2ProtobufStruct(t *testing.T) {
	s, err := Struct2ProtobufStruct(sampleStruct{Name: "alice", Age: 30})
	if assert.NoError(t, err) {
		assert.Equal(t, "alice", s.Fields["Name"].GetStringValue())
		assert.Equal(t, float64(30), s.Fields["Age"].GetNumberValue())
	}
}

// === Defensive fallthrough at elabValue (util.go:63) ===

// TestStruct2Map_UnknownValueKind covers the fallthrough branch in elabValue
// for a *pb.Value with no Kind set. Before the bug fix at util.go:63 this
// test FAILED because the function returned the error in the value position
// (with a nil error), so Struct2Map silently produced a map containing the
// error and no error itself. After the fix it correctly returns (nil, error).
func TestStruct2Map_UnknownValueKind(t *testing.T) {
	s := &pb.Struct{Fields: map[string]*pb.Value{
		"x": {}, // nil Kind: doesn't match any of the type assertions
	}}
	_, err := Struct2Map(s)
	assert.Error(t, err)
}

// === Round-trip ===

func TestRoundTrip(t *testing.T) {
	original := map[string]interface{}{
		"str":  "hello",
		"num":  42.0,
		"bool": true,
		"list": []interface{}{"a", "b"},
		"nested": map[string]interface{}{
			"k": "v",
		},
	}
	pbStruct, err := Map2Struct(original)
	if !assert.NoError(t, err) {
		return
	}
	roundTripped, err := Struct2Map(pbStruct)
	if assert.NoError(t, err) {
		assert.Equal(t, "hello", roundTripped["str"])
		assert.Equal(t, 42.0, roundTripped["num"])
		assert.Equal(t, true, roundTripped["bool"])

		list, ok := roundTripped["list"].([]interface{})
		if assert.True(t, ok) && assert.Len(t, list, 2) {
			assert.Equal(t, "a", list[0])
			assert.Equal(t, "b", list[1])
		}

		nested, ok := roundTripped["nested"].(map[string]interface{})
		if assert.True(t, ok) {
			assert.Equal(t, "v", nested["k"])
		}
	}
}
