package common

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"gorm.io/gorm/schema"
)

func SortNameToDBName(sortName string, model interface{}) string {
	var typ reflect.Type
	if reflect.ValueOf(model).Kind() == reflect.Ptr {
		typ = reflect.Indirect(reflect.ValueOf(model)).Type()
	} else {
		typ = reflect.ValueOf(model).Type()
	}

	//log.Println(sortName, typ.Kind())
	if typ.Kind() != reflect.Struct {
		return ""
	}

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

func NSortNameToDBName(sortName string, model interface{}) (dbName string) {
	//modify sortName
	namesChain := strings.Split(sortName, ".")

	for _, n := range namesChain {
		f, name := searchField(n, model)
		if f == nil {
			return ""
		}
		model = f
		dbName += name + "."
	}
	return strings.TrimRight(dbName, ".")
}

func searchField(name string, model interface{}) (field interface{}, n string) {
	name = strings.ToLower(name)

	var typ reflect.Type
	var val reflect.Value

	if reflect.ValueOf(model).Kind() == reflect.Ptr {
		typ = reflect.Indirect(reflect.ValueOf(model)).Type()
		val = reflect.Indirect(reflect.ValueOf(model))
	} else {
		typ = reflect.ValueOf(model).Type()
		val = reflect.ValueOf(model)
	}

	//log.Println(name, typ.Kind(), typ.Name(), typ.String())

	if typ.Kind() != reflect.Struct {
		log.Printf("Not struct: %v\n", typ.Name())
		return nil, ""
	}

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Type.Kind() == reflect.Struct && f.Anonymous {
			f, n := searchField(name, val.Field(i).Interface())
			if n != "" {
				return f, n
			}
		}
		if sName, dbNme := fieldName(f); name == sName {
			return val.Field(i).Interface(), dbNme
		}
	}
	log.Printf("Not found field %s in struct %v\n", name, typ.Name())
	return nil, ""
}

func fieldName(f reflect.StructField) (sortName, dbName string) {
	if t := f.Tag.Get("cursor"); t != "" {
		sortName = t
	} else {
		sortName = strings.ToLower(f.Name)
	}
	dbName = getDbName(f)

	return
}

func getDbName(f reflect.StructField) (dbName string) {
	if f.Type.Kind() == reflect.Struct {
		return fmt.Sprintf(`"%s"`, f.Name)
	}

	field := (&schema.Schema{}).ParseField(f)
	if field != nil && field.DBName != "" {
		dbName = field.DBName
	}
	dbName = (&schema.NamingStrategy{}).ColumnName("", field.Name)

	return
}
