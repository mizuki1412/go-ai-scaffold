package arraykit

import "testing"

func TestListItemValues(t *testing.T) {
	list := []map[string]any{
		{"id": 1, "profile": map[string]any{"age": 20}},
		{"id": 2, "profile": map[string]any{"age": 30}},
	}
	got := ListItemValues(list, "id")
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("提取列值不符: %v", got)
	}

	ages := ListItemValues(list, "profile", "age")
	if len(ages) != 2 || ages[0] != 20 || ages[1] != 30 {
		t.Fatalf("subKey 提取不符: %v", ages)
	}

	unique := ListItemValuesUnique([]map[string]any{
		{"k": "a"}, {"k": "a"}, {"k": "b"},
	}, "k")
	if len(unique) != 2 {
		t.Fatalf("去重结果不符: %v", unique)
	}
}

func TestItemValueStruct(t *testing.T) {
	type user struct {
		Name string
		Age  int
	}
	v, ok := ItemValue(user{Name: "tom"}, "Name")
	if !ok || v != "tom" {
		t.Fatalf("struct 取值不符: %v %v", v, ok)
	}
}
