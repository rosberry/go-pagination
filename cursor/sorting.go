package cursor

import (
	"reflect"
	"strings"

	"github.com/rosberry/go-pagination/common"
)

type (
	sortingElem struct {
		Field     string `json:"field" form:"field"`
		Direction string `json:"direction" form:"direction"`
	}

	sorting []sortingElem
)

func (srt *sorting) toCursor(model interface{}) *Cursor {
	if srt == nil {
		return nil
	}

	cursor := &Cursor{
		Limit:    common.DefaultLimit,
		Backward: false,
	}

	for _, e := range *srt {
		direction, ok := common.DirectionByString[strings.ToLower(e.Direction)]
		if !ok {
			direction = common.DirectionAsc
		}

		fieldName := common.NSortNameToDBName(e.Field, model)
		if fieldName == "" {
			return nil
		}

		cursor.AddField(fieldName, nil, direction)
	}

	// check and add id field
	var idExist bool

	for _, f := range cursor.Fields {
		if f.Name == "id" {
			idExist = true
			break
		}
	}

	if !idExist {
		var typ reflect.Type
		if reflect.ValueOf(model).Kind() == reflect.Ptr {
			typ = reflect.Indirect(reflect.ValueOf(model)).Type()
		} else {
			typ = reflect.ValueOf(model).Type()
		}

		for i := 0; i < typ.NumField(); i++ {
			structField := typ.Field(i)
			if strings.ToLower(structField.Name) == "id" {
				cursor.AddField("id", nil, common.DirectionAsc)
				break
			}
		}
	}

	return cursor
}
