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
		Next    string `json:"next"`
		Prev    string `json:"prev"`
		HasNext bool   `json:"hasNext"`
		HasPrev bool   `json:"hasPrev"`
	}

	ScopeFunc func(db *gorm.DB) *gorm.DB

	InitDecode struct {
		model reflect.Value
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

//Decode request to cursor
func (d *InitDecode) Decode(c *gin.Context, cg DefaultCursorGetter) *Cursor {
	sortingQuery := c.Query("sorting")
	cursorQuery := c.Query("cursor")

	return decodeAction(d, sortingQuery, cursorQuery, cg)
}

func decodeAction(d *InitDecode, sortingQuery, cursorQuery string, cg DefaultCursorGetter) *Cursor {
	if cursorQuery != "" {
		//Work with cursor
		//Decode string to cursor
		cursor := decodeCursor(cursorQuery)
		if cursor != nil {
			return cursor
		}
	}

	if sortingQuery != "" {
		var sort sorting
		err := json.Unmarshal([]byte(sortingQuery), &sort)
		if err != nil {
			log.Println(err)
		}
		cursor := sort.toCursor(d.model)
		if cursor != nil {
			return cursor
		}
	}

	//Make default cursor
	cursor := cg()

	//cursor to DB
	//return *gorm.DB
	return cursor
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
