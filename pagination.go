package pagination

import (
	"log"
	"pagination/cursor"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type (
	Paginator struct {
		Options
		PageInfo *PageInfo

		cursor *cursor.Cursor
	}

	Options struct {
		GinContext    *gin.Context
		DefaultCursor *cursor.Cursor
		Limit         uint
	}

	PageInfo struct {
		Next      string `json:"next"`
		Prev      string `json:"prev"`
		HasNext   bool   `json:"hasNext"`
		HasPrev   bool   `json:"hasPrev"`
		TotalRows int    `json:"totalRows"`
	}
)

func New(o Options) (*Paginator, error) {
	log.Println("New paginator!")
	return &Paginator{Options: o}, nil
}

func (p *Paginator) Find(tx *gorm.DB, dst interface{}) error {
	//check what dst is pointer
	//execute query
	err := tx.Find(dst).Error //TODO:with scope
	if err != nil {
		return err
	}

	//calc pagination info
	//save paginationInfo to p
	return nil
}

func (p *Paginator) Decode(c *gin.Context) (*cursor.Cursor, error) {
	return &cursor.Cursor{}, nil
}
