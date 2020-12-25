package pagination

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"reflect"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type (
	//PaginationResponse struct
	PaginationResponse struct {
		Next      string `json:"next"`
		Prev      string `json:"prev"`
		HasNext   bool   `json:"hasNext"`
		HasPrev   bool   `json:"hasPrev"`
		TotalRows int    `json:"totalRows,omitempty"`
	}

	ScopeFunc func(db *gorm.DB) *gorm.DB

	InitDecode struct {
		model reflect.Value
		limit int
		db    *gorm.DB
	}
)

//Model init for cursor decode
func Model(model interface{}) *InitDecode {
	if reflect.ValueOf(model).Kind() == reflect.Ptr {
		return &InitDecode{
			model: reflect.Indirect(reflect.ValueOf(model)),
		}
	}
	return &InitDecode{
		model: reflect.ValueOf(model),
	}
}

//Limit set global limit for cursor
//This value overrides whatever specify in DefaultCursorGetter
func (d *InitDecode) Limit(limit int) *InitDecode {
	if limit <= 0 {
		limit = defaultLimit
	}
	d.limit = limit
	return d
}

//DB set database connection
func (d *InitDecode) DB(db *gorm.DB) *InitDecode {
	if db != nil {
		d.db = db
	}
	return d
}

//Decode request to cursor
func (d *InitDecode) Decode(c *gin.Context, cg DefaultCursorGetter) (*Cursor, error) {
	sortingQuery := c.Query("sorting")
	cursorQuery := c.Query("cursor")

	return decodeAction(d, sortingQuery, cursorQuery, cg)
}

func decodeAction(d *InitDecode, sortingQuery, cursorQuery string, cg DefaultCursorGetter) (*Cursor, error) {
	if cursorQuery != "" && sortingQuery != "" {
		return nil, ErrCursorAndSortingTogether
	}

	var cursor *Cursor
	if cursorQuery != "" {
		//Work with cursor
		//Decode string to cursor
		cursor = decodeCursor(cursorQuery)
		if cursor == nil {
			return nil, ErrInvalidCursor
		}

	} else if sortingQuery != "" {
		var sort sorting
		err := json.Unmarshal([]byte(sortingQuery), &sort)
		if err != nil {
			log.Println("Unmarshal err:", err)
			return nil, ErrInvalidSorting
		}
		cursor = sort.toCursor(d.model)
		if cursor == nil {
			return nil, ErrInvalidSorting
		}

	} else {
		//Make default cursor
		cursor = cg()
		if cursor == nil {
			return nil, ErrInvalidDefaultCursor
		}
	}

	if d.limit > 0 {
		cursor.Limit = d.limit
	}
	if d.db != nil {
		cursor.db = d.db
	}

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
