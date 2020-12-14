package pagination

import (
	"reflect"
	"strings"
)

type (
	sortingElem struct {
		Field     string `json:"field" form:"field"`
		Direction string `json:"direction" form:"direction"`
	}

	sorting []sortingElem
)

func (srt *sorting) toCursor(model reflect.Value) *Cursor {
	if srt == nil {
		return nil
	}

	typ := model.Type()
	cursor := &Cursor{
		Limit:    defaultLimit,
		Backward: false,
	}

	for _, e := range *srt {
		direction, ok := DirectionByString[strings.ToLower(e.Direction)]
		if !ok {
			direction = DirectionAsc
		}
		fieldName := sortNameToDBName(e.Field, typ)
		cursor.AddField(fieldName, nil, direction)
	}

	//check and add id field
	var idExist bool
	for _, f := range cursor.Fields {
		if f.Name == "id" {
			idExist = true
			break
		}
	}

	if !idExist {
		for i := 0; i < typ.NumField(); i++ {
			structField := typ.Field(i)
			if strings.ToLower(structField.Name) == "id" {
				cursor.AddField("id", nil, DirectionAsc)
				break
			}
		}
	}

	return cursor
}
