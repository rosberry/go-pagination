package pagination

import (
	"encoding/base64"
	"encoding/json"
	"log"

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
)

//Decode request to cursor
func Decode(c *gin.Context, cg DefaultCursorGetter) *Cursor {
	sortingQuery := c.Query("sorting")
	cursorQuery := c.Query("cursor")

	if cursorQuery != "" {
		//Work with cursor
		//Decode string to cursor
		cursor := decodeCursor(cursorQuery)
		if cursor != nil {
			return cursor
		}
	}

	if sortingQuery != "" {
		log.Println("Need unmarshal: ", sortingQuery)
		var sort sorting
		err := json.Unmarshal([]byte(sortingQuery), &sort)
		if err != nil {
			log.Println(err)
		}
		cursor := sort.toCursor()
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

//decodeCursor from base64 string
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
