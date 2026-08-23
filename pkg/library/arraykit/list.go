package arraykit

// 从列表提取列值（移植自 GoFrame util/gutil，MIT）。

import "reflect"

// ListItemValues 从切片中每个 map/struct 元素提取 key 对应的值组成新切片。
// list 支持 []map[string]any、[]struct（含指针），subKey 用于取嵌套子值。
func ListItemValues(list any, key any, subKey ...any) (values []any) {
	var reflectValue reflect.Value
	if v, ok := list.(reflect.Value); ok {
		reflectValue = v
	} else {
		reflectValue = reflect.ValueOf(list)
	}
	reflectKind := reflectValue.Kind()
	for reflectKind == reflect.Pointer {
		reflectValue = reflectValue.Elem()
		reflectKind = reflectValue.Kind()
	}
	switch reflectKind {
	case reflect.Slice, reflect.Array:
		if reflectValue.Len() == 0 {
			return
		}
		values = []any{}
		for i := 0; i < reflectValue.Len(); i++ {
			if value, ok := ItemValue(reflectValue.Index(i), key); ok {
				if len(subKey) > 0 && subKey[0] != nil {
					if subValue, ok := ItemValue(value, subKey[0]); ok {
						value = subValue
					} else {
						continue
					}
				}
				if array, ok := value.([]any); ok {
					values = append(values, array...)
				} else {
					values = append(values, value)
				}
			}
		}
	}
	return
}

// ListItemValuesUnique 同 ListItemValues，并对结果去重。
func ListItemValuesUnique(list any, key string, subKey ...any) []any {
	values := ListItemValues(list, key, subKey...)
	if len(values) > 0 {
		m := make(map[any]struct{}, len(values))
		for i := 0; i < len(values); {
			value := values[i]
			if t, ok := value.([]byte); ok {
				value = string(t)
			}
			if _, ok := m[value]; ok {
				values = append(values[:i], values[i+1:]...)
			} else {
				m[value] = struct{}{}
				i++
			}
		}
	}
	return values
}

// ItemValue 取 item（map/*map/struct/*struct/slice）中 key 指定的属性值。
func ItemValue(item any, key any) (value any, found bool) {
	var reflectValue reflect.Value
	if v, ok := item.(reflect.Value); ok {
		reflectValue = v
	} else {
		reflectValue = reflect.ValueOf(item)
	}
	reflectKind := reflectValue.Kind()
	if reflectKind == reflect.Interface {
		reflectValue = reflectValue.Elem()
		reflectKind = reflectValue.Kind()
	}
	for reflectKind == reflect.Pointer {
		reflectValue = reflectValue.Elem()
		reflectKind = reflectValue.Kind()
	}
	keyValue := reflect.ValueOf(key)
	switch reflectKind {
	case reflect.Array, reflect.Slice:
		// key 必须为 string，按元素字段继续提取
		values := ListItemValues(reflectValue, keyValue.String())
		if values == nil {
			return nil, false
		}
		return values, true

	case reflect.Map:
		v := reflectValue.MapIndex(keyValue)
		if v.IsValid() {
			found = true
			value = v.Interface()
		}

	case reflect.Struct:
		v := reflectValue.FieldByName(keyValue.String())
		if v.IsValid() {
			found = true
			value = v.Interface()
		}
	}
	return
}
