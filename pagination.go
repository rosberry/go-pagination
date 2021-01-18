package pagination

import (
	"log"
	"reflect"

	"github.com/rosberry/go-pagination/common"
	"github.com/rosberry/go-pagination/cursor"

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
		DB            *gorm.DB
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

	//err := p.DB.Table("(?) as t", tx.Session(&gorm.Session{})).Scopes(p.cursor.Scope()).Find(dst).Error
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

	//save paginationInfo to p
	pageInfo := &PageInfo{
		Next:      nextCursor.Encode(),
		Prev:      prevCursor.Encode(),
		HasNext:   p.checkPage(tx.Session(&gorm.Session{}), nextCursor.Scope()),
		HasPrev:   p.checkPage(tx.Session(&gorm.Session{}), prevCursor.Scope()),
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

	//cursor.DB = p.DB
	p.cursor = cursor
	return nil
}

func count(tx *gorm.DB) (count int64) {
	if err := tx.Session(&gorm.Session{}).Select("count(1)").Limit(1).Count(&count).Error; err != nil {
		log.Println(err)
		return -1
	}
	return
}

func (p *Paginator) checkPage(tx *gorm.DB, scope func(*gorm.DB) *gorm.DB) (isExist bool) {
	var count int64
	if err := p.DB.Table("(?) as t", tx.Session(&gorm.Session{})).Scopes(scope).Select("count(1)").Limit(1).Count(&count).Error; err != nil {
		log.Println(err)
		return
	}
	if count > 0 {
		return true
	}
	return
}
