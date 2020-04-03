package sqlex

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"
)

type structFields map[string]*reflect.Value

func buildFields(tv *reflect.Value, tag string, sep string) (structFields, error) {
	if tv.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("%s should be ptr but isn't", tv.String())
	}
	if tv.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s should be ptr of struct but isn't", tv.String())
	}

	typ := tv.Type()
	fiels := make(structFields)
	for i := 0; i < typ.Elem().NumField(); i++ {
		field := tv.Elem().Field(i)
		fieldType := typ.Elem().Field(i)
		if !ast.IsExported(fieldType.Name) {
			continue
		}
		name := fieldType.Name
		if tag != "" {
			tagStr := fieldType.Tag.Get(tag)
			if tagStr != "" {
				if tagStr == "-" {
					continue
				}
				if sep != "" {
					tagStr = strings.Split(tagStr, sep)[0]
				}
				name = tagStr
			}
		}
		fiels[name] = &field
	}
	return fiels, nil
}
