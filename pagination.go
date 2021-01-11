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
	p.decode()
	//check what dst is pointer to slice
	if reflect.ValueOf(dst).Kind() != reflect.Ptr {
		return common.ErrInvalidFindDestinationNotPointer
	}
	if reflect.Indirect(reflect.ValueOf(dst)).Kind() != reflect.Slice {
		return common.ErrInvalidFindDestinationNotSlice
	}

	//execute query
	if p.cursor == nil {
		return common.ErrInvalidCursor
	}

	err := tx.Session(&gorm.Session{}).Scopes(p.cursor.Scope()).Find(dst).Error
	if err != nil {
		return err
	}

	//calc paginationinfo
	p.PageInfo = p.calcPageInfo(tx, dst)
	return nil
}

func (p *Paginator) calcPageInfo(tx *gorm.DB, dst interface{}) *PageInfo {
	//query for totalRow
	totalRows := count(tx.Session(&gorm.Session{}))

	object := reflect.Indirect(reflect.ValueOf(dst))
	//first elem to prevCursor
	nextCursor := p.cursor.ToCursor(object.Index(object.Len() - 1))

	//last elem to nextCursor
	prevCursor := p.cursor.ToCursor(object.Index(0))
	prevCursor.Backward = true

	//query for hasPrev
	var prevCnt int64
	tx.Session(&gorm.Session{}).Scopes(prevCursor.Scope()).Count(&prevCnt)

	//query for hasNext
	var nextCnt int64
	tx.Session(&gorm.Session{}).Scopes(nextCursor.Scope()).Count(&nextCnt)

	//save paginationInfo to p
	pageInfo := &PageInfo{
		Next:      nextCursor.Encode(),
		Prev:      prevCursor.Encode(),
		HasNext:   nextCnt > 0,
		HasPrev:   prevCnt > 0,
		TotalRows: int(totalRows),
	}

	return pageInfo
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
