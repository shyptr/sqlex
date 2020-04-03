package sqlex

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
)

func SelectOne(runner BaseRunner, strct interface{}, selectBuilder SelectBuilder) (int, error) {
	query, err := selectBuilder.RunWith(runner).Query()
	if err != nil {
		return 0, err
	}
	defer query.Close()
	if query.Next() {
		strctVal := reflect.ValueOf(strct)
		if err = buildOne(&strctVal, query, "", ""); err != nil {
			return 0, err
		}
	}
	return 0, nil
}

func SelectOneByTag(runner BaseRunner, strct interface{}, selectBuilder SelectBuilder, tag string, sep string) (int, error) {
	query, err := selectBuilder.RunWith(runner).Query()
	if err != nil {
		return 0, err
	}
	defer query.Close()
	if query.Next() {
		strctVal := reflect.ValueOf(strct)
		if err = buildOne(&strctVal, query, tag, sep); err != nil {
			return 0, err
		}
	}
	return 1, nil
}

func SelectList(runner BaseRunner, slice interface{}, selectBuilder SelectBuilder) (int, error) {
	query, err := selectBuilder.RunWith(runner).Query()
	if err != nil {
		return 0, err
	}
	defer query.Close()
	sliceVal := reflect.ValueOf(slice)
	if sliceVal.Kind() != reflect.Ptr {
		return 0, fmt.Errorf("%s should be slice's ptr but it isn't", sliceVal.String())
	}
	if sliceVal.IsNil() {
		sliceVal.Set(reflect.New(sliceVal.Type().Elem()))
	}
	sliceVal = sliceVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return 0, fmt.Errorf("%s should be slice's ptr but it isn't", sliceVal.String())
	}
	typ := sliceVal.Type().Elem()
	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() != reflect.Struct {
		return 0, fmt.Errorf("%s should be slice's ptr of struct's ptr but it isn't", sliceVal.String())
	} else if typ.Kind() != reflect.Struct {
		return 0, fmt.Errorf("%s shoule be slice's ptr of struct but it isn't", sliceVal.String())
	}
	newSlice := reflect.MakeSlice(sliceVal.Type(), 0, 16)
	for query.Next() {
		strctVal := reflect.New(typ)
		if typ.Kind() == reflect.Ptr {
			strctVal = strctVal.Elem()
			if strctVal.IsNil() {
				strctVal.Set(reflect.New(typ.Elem()))
			}
		}
		if err = buildOne(&strctVal, query, "", ""); err != nil {
			return 0, err
		}
		if typ.Kind() == reflect.Struct {
			newSlice = reflect.Append(newSlice, strctVal.Elem())
		} else {
			newSlice = reflect.Append(newSlice, strctVal)
		}
	}
	sliceVal.Set(newSlice)
	return sliceVal.Len(), nil
}

func SelectListByTag(runner BaseRunner, slice interface{}, selectBuilder SelectBuilder, tag string, sep string) (int, error) {
	query, err := selectBuilder.RunWith(runner).Query()
	if err != nil {
		return 0, err
	}
	defer query.Close()
	sliceVal := reflect.ValueOf(slice)
	if sliceVal.Kind() != reflect.Ptr {
		return 0, fmt.Errorf("%s should be slice's ptr but it isn't", sliceVal.String())
	}
	if sliceVal.IsNil() {
		sliceVal.Set(reflect.New(sliceVal.Type().Elem()))
	}
	sliceVal = sliceVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return 0, fmt.Errorf("%s should be slice's ptr but it isn't", sliceVal.String())
	}
	typ := sliceVal.Type().Elem()
	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() != reflect.Struct {
		return 0, fmt.Errorf("%s should be slice's ptr of struct's ptr but it isn't", sliceVal.String())
	} else if typ.Kind() != reflect.Struct {
		return 0, fmt.Errorf("%s shoule be slice's ptr of struct but it isn't", sliceVal.String())
	}
	newSlice := reflect.MakeSlice(sliceVal.Type(), 0, 16)
	for query.Next() {
		strctVal := reflect.New(typ)
		if typ.Kind() == reflect.Ptr {
			strctVal = strctVal.Elem()
			if strctVal.IsNil() {
				strctVal.Set(reflect.New(typ.Elem()))
			}
		}
		if err = buildOne(&strctVal, query, tag, sep); err != nil {
			return 0, err
		}
		if typ.Kind() == reflect.Struct {
			newSlice = reflect.Append(newSlice, strctVal.Elem())
		} else {
			newSlice = reflect.Append(newSlice, strctVal)
		}
	}
	sliceVal.Set(newSlice)
	return sliceVal.Len(), nil
}

func buildOne(strctVal *reflect.Value, query *sql.Rows, tag string, sep string) error {
	fields, err := buildFields(strctVal, tag, sep)
	if err != nil {
		return err
	}
	cols, err := query.Columns()
	if err != nil {
		return err
	}
	columnTypes, err := query.ColumnTypes()
	if err != nil {
		return err
	}
	dest := make([]interface{}, len(columnTypes))
	for index, scanType := range columnTypes {
		switch scanType.ScanType().String() {
		case "string", "interface {}":
			dest[index] = &sql.NullString{}
		case "bool":
			dest[index] = &sql.NullBool{}
		case "float64":
			dest[index] = &sql.NullFloat64{}
		case "int32":
			dest[index] = &sql.NullInt32{}
		case "int64":
			dest[index] = &sql.NullInt64{}
		case "time.Time":
			dest[index] = &sql.NullTime{}
		default:
			dest[index] = reflect.New(scanType.ScanType()).Interface()
		}
	}
	err = query.Scan(dest...)
	if err != nil {
		return err
	}
	for index, col := range columnTypes {
		fptr, ok := fields[cols[index]]
		if !ok {
			continue
		}
		f := *fptr
		switch val := dest[index].(type) {
		case driver.Valuer:
			var value interface{}
			switch col.ScanType().String() {
			case "string", "interface {}":
				value = dest[index].(*sql.NullString).String
			case "bool":
				value = dest[index].(*sql.NullBool).Bool
			case "float64":
				value = dest[index].(*sql.NullFloat64).Float64
			case "int32":
				value = dest[index].(*sql.NullInt32).Int32
			case "int64":
				value = dest[index].(*sql.NullInt64).Int64
			case "time.Time":
				value = dest[index].(*sql.NullTime).Time
			}
			ftyp := f.Type()
			if f.Kind() == reflect.Ptr {
				if f.IsNil() {
					f.Set(reflect.New(f.Type().Elem()))
				}
				ftyp = f.Type().Elem()
				f = f.Elem()
			}
			valueVal := reflect.ValueOf(value)
			if !valueVal.Type().ConvertibleTo(ftyp) {
				continue
			}
			f.Set(valueVal.Convert(ftyp))
		default:
			valValue := reflect.ValueOf(val)
			if f.Kind() != reflect.Ptr {
				if valValue.IsNil() {
					valValue.Set(reflect.New(valValue.Type().Elem()))
				}
				valValue = valValue.Elem()
			}
			if !valValue.Type().ConvertibleTo(f.Type()) {
				continue
			}
			f.Set(valValue.Convert(f.Type()))
		}
	}
	return nil
}
