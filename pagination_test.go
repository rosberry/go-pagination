package pagination

import (
	"reflect"
	"testing"
)

var (
	emptyCursorString  = ""
	failedCursorString = "abc"
	notCursorJSON      = `{"a":1,"b":2"}`
)

func TestDecodeCursor(t *testing.T) {
	result := decodeCursor(emptyCursorString)
	if result != nil {
		t.Errorf("Result must be %v for string: %v", "nil", emptyCursorString)
	}

	result = decodeCursor(failedCursorString)
	if result != nil {
		t.Errorf("Result must be %v for string: %v", "nil", failedCursorString)
	}

	result = decodeCursor(notCursorJSON)
	if result != nil {
		t.Errorf("Result must be %v for string: %v", "nil", notCursorJSON)
	}

	result = decodeCursor(defaultCursorEncodeBase64Str)
	if !reflect.DeepEqual(defaultCursor, result) {
		t.Error("Result must be equal defaultCursor")
	}
}
