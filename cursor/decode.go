package cursor

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"pagination/common"
)

func DecodeAction(sortingQuery, cursorQuery string, defaultCursor *Cursor, model interface{}, limit uint) (*Cursor, error) {
	if cursorQuery != "" && sortingQuery != "" {
		return nil, common.ErrCursorAndSortingTogether
	}

	var cursor *Cursor
	if cursorQuery != "" {
		//Work with cursor
		//Decode string to cursor
		cursor = decodeCursor(cursorQuery)
		if cursor == nil {
			return nil, common.ErrInvalidCursor
		}

	} else if sortingQuery != "" {
		var sort sorting
		err := json.Unmarshal([]byte(sortingQuery), &sort)
		if err != nil {
			return nil, common.ErrInvalidSorting
		}
		cursor = sort.toCursor(model)
		if cursor == nil {
			return nil, common.ErrInvalidSorting
		}

	} else {
		//Make default cursor
		cursor = defaultCursor
		if cursor == nil {
			return nil, common.ErrInvalidDefaultCursor
		}
	}

	if limit > 0 {
		cursor.Limit = int(limit)
	}

	/*
		if d.db != nil {
			cursor.db = d.db
		}
	*/

	return cursor, nil
}

//decodeCursorString - decode cursor from base64 string
func decodeCursor(s string) *Cursor {
	var cursor Cursor
	// Decode
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Println("Decode err:", err)
		return nil
	}

	err = json.Unmarshal([]byte(raw), &cursor)
	if err != nil {
		log.Println("Unmarshal err:", err)
		return nil
	}

	log.Printf("Cursor: %+v\n", cursor)
	return &cursor
}
