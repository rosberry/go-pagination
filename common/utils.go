package common

import (
	"reflect"

	"gorm.io/gorm/schema"
)

func SortNameToDBName(sortName string, model interface{}) string {
	typ := reflect.ValueOf(model).Type()
	for i := 0; i < typ.NumField(); i++ {
		structField := typ.Field(i)
		name := columnName(structField)

		if sortName == name {
			field := (&schema.Schema{}).ParseField(structField)
			if field.DBName != "" {
				return field.DBName
			}
			return (&schema.NamingStrategy{}).ColumnName("", field.Name)
		}
	}

	return ""
}

func columnName(field reflect.StructField) string {
	tags := field.Tag
	var colName string
	colName = tags.Get("cursor")
	if colName == "" {
		colName = (&schema.NamingStrategy{}).ColumnName("", field.Name)
	}
	return colName
}

//Backward of order type
func (dt DirectionType) Backward(ok bool) DirectionType {
	if !ok {
		return dt
	}
	switch dt {
	case DirectionDesc:
		return DirectionAsc
	case DirectionAsc:
		return DirectionDesc
	default:
		return DirectionAsc
	}
}
