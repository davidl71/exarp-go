package tools

import (
	"encoding/json"
	"testing"
)

func TestParamInt_JSONNumberAndInt(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(`{"limit":10,"a":10.0,"b":"12"}`), &params); err != nil {
		t.Fatal(err)
	}
	if got := ParamInt(params, "limit", 0); got != 10 {
		t.Fatalf("limit: got %d want 10", got)
	}
	if got := ParamInt(params, "a", 0); got != 10 {
		t.Fatalf("a: got %d want 10", got)
	}
	if got := ParamInt(params, "b", 0); got != 12 {
		t.Fatalf("b: got %d want 12", got)
	}
	if got := ParamInt(params, "missing", 99); got != 99 {
		t.Fatalf("missing: got %d want 99", got)
	}
}

func TestParamFloat64OK(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(`{"x":0.25,"y":"0.5","z":"nope"}`), &params); err != nil {
		t.Fatal(err)
	}
	if v, ok := ParamFloat64OK(params, "x"); !ok || v != 0.25 {
		t.Fatalf("x: ok=%v v=%v", ok, v)
	}
	if v, ok := ParamFloat64OK(params, "y"); !ok || v != 0.5 {
		t.Fatalf("y: ok=%v v=%v", ok, v)
	}
	if _, ok := ParamFloat64OK(params, "z"); ok {
		t.Fatal("z should not convert")
	}
	if _, ok := ParamFloat64OK(params, "absent"); ok {
		t.Fatal("absent should be false")
	}
}

func TestParamStringSlice_JSONArray(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(`{"ids":["a","b"]}`), &params); err != nil {
		t.Fatal(err)
	}
	got := ParamStringSlice(params, "ids")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %#v", got)
	}
	if ParamStringSlice(params, "missing") != nil {
		t.Fatal("missing should be nil")
	}
}

func TestParamStringSliceTrimmed(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(`{"tags":[" a ","","b"],"empties":["  ","\t"]}`), &params); err != nil {
		t.Fatal(err)
	}
	got := ParamStringSliceTrimmed(params, "tags")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("got %#v", got)
	}
	if ParamStringSliceTrimmed(params, "missing") != nil {
		t.Fatal("missing should be nil")
	}
	if ParamStringSliceTrimmed(params, "empties") != nil {
		t.Fatal("all whitespace should be nil")
	}
}

func TestParamTaskDependencyIDs_nestedAndJSONText(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(`{"dependencies":[["T-100","T-101"]]}`), &params); err != nil {
		t.Fatal(err)
	}
	got := ParamTaskDependencyIDs(params, "dependencies")
	if len(got) != 2 || got[0] != "T-100" || got[1] != "T-101" {
		t.Fatalf("nested array: got %#v", got)
	}
	if err := json.Unmarshal([]byte(`{"dependencies":"[\"T-200\",\"T-201\"]"}`), &params); err != nil {
		t.Fatal(err)
	}
	got = ParamTaskDependencyIDs(params, "dependencies")
	if len(got) != 2 || got[0] != "T-200" || got[1] != "T-201" {
		t.Fatalf("JSON text blob: got %#v", got)
	}
	// Single string token that older parsers treated as one malformed dependency (brackets included).
	params = map[string]interface{}{"dependencies": `["T-1774817606330858000"]`}
	got = ParamTaskDependencyIDs(params, "dependencies")
	if len(got) != 1 || got[0] != "T-1774817606330858000" {
		t.Fatalf("JSON array string value: got %#v", got)
	}
	if ParamTaskDependencyIDs(map[string]interface{}{}, "dependencies") != nil {
		t.Fatal("missing key should be nil")
	}
}

func TestParamStringSliceTrimmedCommaSeparated(t *testing.T) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(`{"ids":" T-1 , T-2 ","arr":["x"," y "]}`), &params); err != nil {
		t.Fatal(err)
	}
	got := ParamStringSliceTrimmedCommaSeparated(params, "ids")
	if len(got) != 2 || got[0] != "T-1" || got[1] != "T-2" {
		t.Fatalf("ids: got %#v", got)
	}
	gotArr := ParamStringSliceTrimmedCommaSeparated(params, "arr")
	if len(gotArr) != 2 || gotArr[0] != "x" || gotArr[1] != "y" {
		t.Fatalf("arr: got %#v", gotArr)
	}
	if err := json.Unmarshal([]byte(`{"tagsJson":"[\"backend\",\"urgent\"]"}`), &params); err != nil {
		t.Fatal(err)
	}
	gotJSONStr := ParamStringSliceTrimmedCommaSeparated(params, "tagsJson")
	if len(gotJSONStr) != 2 || gotJSONStr[0] != "backend" || gotJSONStr[1] != "urgent" {
		t.Fatalf("tagsJson string: got %#v", gotJSONStr)
	}
	if err := json.Unmarshal([]byte(`{"emptyArr":"[]"}`), &params); err != nil {
		t.Fatal(err)
	}
	if ParamStringSliceTrimmedCommaSeparated(params, "emptyArr") != nil {
		t.Fatal("empty JSON array string should yield nil tags")
	}
	if ParamStringSliceTrimmedCommaSeparated(params, "missing") != nil {
		t.Fatal("missing should be nil")
	}
}

func TestParamIntOK(t *testing.T) {
	params := map[string]interface{}{"n": float64(3)}
	if v, ok := ParamIntOK(params, "n"); !ok || v != 3 {
		t.Fatalf("got %d ok=%v", v, ok)
	}
	if _, ok := ParamIntOK(params, "missing"); ok {
		t.Fatal("want false")
	}
}
