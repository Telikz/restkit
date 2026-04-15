package utils

import "reflect"

// UpdateFields takes the existing entity and applies it first, then uses request overrides
// for any non-nil pointer fields in the request. This allows for partial updates without losing existing data.
// T is the destination type, E is the existing entity type, and R is the request type.
func UpdateFields[T any, E any, R any](base T, existing E, req R) T {
	result := base
	rv := reflect.ValueOf(&result).Elem()

	// First copy from existing entity
	ev := reflect.ValueOf(existing)
	if ev.Kind() == reflect.Pointer {
		ev = ev.Elem()
	}

	for i := 0; i < rv.NumField(); i++ {
		fieldName := rv.Type().Field(i).Name
		if existingField := ev.FieldByName(fieldName); existingField.IsValid() {
			rv.Field(i).Set(existingField)
		}
	}

	// Then apply request overrides for non-nil pointers
	sv := reflect.ValueOf(req)
	if sv.Kind() == reflect.Pointer {
		sv = sv.Elem()
	}

	for i := 0; i < sv.NumField(); i++ {
		field := sv.Field(i)
		fieldName := sv.Type().Field(i).Name

		if field.Kind() == reflect.Pointer && !field.IsNil() {
			if destField := rv.FieldByName(fieldName); destField.IsValid() && destField.CanSet() {
				destField.Set(field.Elem())
			}
		}
	}

	return result
}
