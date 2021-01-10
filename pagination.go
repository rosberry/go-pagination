package pagination

import (
	"log"
	"pagination/common"
	"pagination/cursor"
	"reflect"

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
		Model         interface{}
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
	if o.DefaultCursor == nil {
		o.DefaultCursor = &cursor.Cursor{
			Fields: []cursor.Field{
				cursor.Field{
					Name:      "id",
					Value:     nil,
					Direction: common.DirectionAsc,
				},
			},
			Limit: common.DefaultLimit,
		}
	}
	return &Paginator{Options: o}, nil
}

func (p *Paginator) Find(tx *gorm.DB, dst interface{}) error {
	log.Println("Paginator: Find()")
	p.decode()
	//check what dst is pointer to slice
	if reflect.ValueOf(dst).Kind() != reflect.Ptr {
		log.Println("dst is not pointer to slice!")
		return common.ErrInvalidFindDestinationNotPointer
	}
	if reflect.Indirect(reflect.ValueOf(dst)).Kind() != reflect.Slice {
		log.Println("dst pointer not to slice!:", reflect.TypeOf(reflect.Indirect(reflect.ValueOf(dst))).Kind())
		return common.ErrInvalidFindDestinationNotSlice
	}

	//query for totalRow
	totalRows := count(tx)

	//execute query
	log.Printf("Paginator (Find): %+v\n", p)
	if p.cursor == nil {
		log.Println("Cursor is nil")
		return common.ErrInvalidCursor
	}
	err := tx.Scopes(p.cursor.Scope()).Find(dst).Error //TODO:with scope
	if err != nil {
		return err
	}

	//calc paginationinfo
	//var pi PageInfo
	object := reflect.Indirect(reflect.ValueOf(dst))
	//first elem to prevCursor
	first := object.Index(0)
	nextCursor := p.cursor.ToCursor(first)
	//last elem to nextCursor
	last := object.Index(object.Len() - 1)
	prevCursor := p.cursor.ToCursor(last)

	//query for hasPrev
	var hasPrev int64
	log.Println("Query for prev")
	tx.Scopes(prevCursor.Scope()).Count(&hasPrev)

	//query for hasNext
	var hasNext int64
	log.Println("Query for next")
	tx.Scopes(nextCursor.Scope()).Count(&hasNext) //!!!Not work!!

	log.Println("nextCursor:", nextCursor)
	log.Println("prevCursor:", prevCursor)
	//save paginationInfo to p
	p.PageInfo = &PageInfo{
		Next:      nextCursor.Encode(),
		Prev:      prevCursor.Encode(),
		HasNext:   hasNext > 0,
		HasPrev:   hasPrev > 0,
		TotalRows: int(totalRows),
	}
	return nil
}

func (p *Paginator) decode() error {
	sortingQuery := p.GinContext.Query("sorting")
	cursorQuery := p.GinContext.Query("cursor")

	cursor, err := cursor.DecodeAction(sortingQuery, cursorQuery, p.DefaultCursor, p.Model, p.Limit)
	if err != nil {
		return err
	}

	p.cursor = cursor
	return nil
}

func count(tx *gorm.DB) (count int64) {
	if err := tx.Count(&count).Error; err != nil {
		log.Println(err)
		return -1
	}
	return
}
