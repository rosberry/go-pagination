package pagination

import (
	"reflect"
	"testing"
)

func TestToCursor(t *testing.T) {
	type User struct {
		ID         uint
		NameOfUser string `gorm:"column:name" json:"name" cursor:"nameOfMyUser"`
		Count      uint
	}

	srting := sorting{
		sortingElem{"nameOfMyUser", "Desc"}, sortingElem{"id", "Asc"}, sortingElem{"count", "Asc"},
	}

	usr := &User{}

	val := reflect.ValueOf(*usr)
	cursor := srting.toCursor(val)

	cursorSuccess := &Cursor{
		Fields: []Field{
			Field{
				Name:      "name",
				Value:     nil,
				Direction: DirectionDesc,
			},
			Field{
				Name:      "id",
				Value:     nil,
				Direction: DirectionAsc,
			},
			Field{
				Name:      "count",
				Value:     nil,
				Direction: DirectionAsc,
			},
		},
		Limit:    defaultLimit,
		Backward: false,
	}

	if !cursorCompare(cursor, cursorSuccess) {
		t.Errorf("Fail sorting to Cursor\n%v\n%v\n", cursor, cursorSuccess)
	}

	//full

	di := Model(usr)
	sortString := `[
		{
			"field": "nameOfMyUser",
			"direction": "desc"
		},
		{
			"field": "id",
			"direction": "Asc"
		},
		{
			"field": "count",
			"direction": "ASC"
		}
	]`
	cursor = decodeAction(di, sortString, "", DefaultCursor)
	if !cursorCompare(cursor, cursorSuccess) {
		t.Errorf("Fail sorting to Cursor\n%v\n%v\n", cursor, cursorSuccess)
	}

}

func cursorCompare(c1, c2 *Cursor) bool {
	if c1.Limit != c2.Limit {
		return false
	}

	if c1.Backward != c2.Backward {
		return false
	}

	if len(c1.Fields) != len(c2.Fields) {
		return false
	}

	for i, f := range c1.Fields {
		if c2.Fields[i] != f {
			return false
		}
	}

	return true
}
