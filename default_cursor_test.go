package pagination

import (
	"reflect"
	"testing"
)

var defaultCursor = &Cursor{
	Fields: []Field{
		Field{
			Name:      "id",
			Value:     nil,
			Direction: DirectionAsc,
		},
	},
	Limit:    defaultLimit,
	Backward: false,
}

func TestDefaultCursor(t *testing.T) {
	cursor := DefaultCursor()

	if !reflect.DeepEqual(defaultCursor, cursor) {
		t.Error("Default cursor failed")
	}
}
