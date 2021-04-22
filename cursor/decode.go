package cursor

import (
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/rosberry/go-pagination/common"
)

func DecodeAction(sortingQuery, cursorQuery, afterQuery, beforeQuery string, defaultCursor *Cursor, model interface{}, limit uint) (cursor, additionalCursor *Cursor, err error) {
	if cursorQuery != "" && sortingQuery != "" {
		return nil, nil, common.ErrCursorAndSortingTogether
	}

	switch {
	case cursorQuery != "":
		// Work with cursor
		// Decode string to cursor
		cursor = decodeCursor(cursorQuery, common.CursorBasic)
		if cursor == nil {
			return nil, nil, common.ErrInvalidCursor
		}
	case afterQuery != "" || beforeQuery != "":
		var afterCursor, beforeCursor *Cursor
		if afterQuery != "" {
			afterCursor = decodeCursor(afterQuery, common.CursorAfter)
			if afterCursor == nil {
				return nil, nil, common.ErrInvalidCursor
			}
		}

		if beforeQuery != "" {
			beforeCursor = decodeCursor(beforeQuery, common.CursorBefore)
			if beforeCursor == nil {
				return nil, nil, common.ErrInvalidCursor
			}
		}

		if afterCursor == nil && beforeCursor != nil {
			return beforeCursor, nil, nil
		}

		return afterCursor, beforeCursor, nil
	case sortingQuery != "":
		var sort sorting

		err := json.Unmarshal([]byte(sortingQuery), &sort)
		if err != nil {
			return nil, nil, common.ErrInvalidSorting
		}

		cursor = sort.toCursor(model)
		if cursor == nil {
			return nil, nil, common.ErrInvalidSorting
		}

		if limit > 0 {
			cursor.Limit = int(limit)
		}
	default:
		// Make default cursor
		cursor = defaultCursor
		if cursor == nil {
			return nil, nil, common.ErrInvalidDefaultCursor
		}

		if limit > 0 {
			cursor.Limit = int(limit)
		}
	}

	return cursor, nil, nil
}

// decodeCursorString - decode cursor from base64 string
func decodeCursor(s string, direction common.CursorDirection) *Cursor {
	var cursor Cursor
	// Decode
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Println("Decode err:", err)
		return nil
	}

	err = json.Unmarshal(raw, &cursor)
	if err != nil {
		log.Println("Unmarshal err:", err)
		return nil
	}

	switch direction {
	case common.CursorAfter:
		cursor.Backward = false
	case common.CursorBefore:
		cursor.Backward = true
	case common.CursorBasic:
	}

	return &cursor
}
