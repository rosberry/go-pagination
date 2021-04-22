package pagination

import (
	"log"
	"reflect"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/rosberry/go-pagination/common"
	"github.com/rosberry/go-pagination/cursor"
)

type (
	Paginator struct {
		options  Options
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
	if o.DefaultCursor == nil {
		o.DefaultCursor = &cursor.Cursor{
			Fields: []cursor.Field{
				{
					Name:      "id",
					Value:     nil,
					Direction: common.DirectionAsc,
				},
			},
			Limit: func() int {
				if o.Limit != 0 {
					return int(o.Limit)
				}
				return common.DefaultLimit
			}(),
		}
	}
	return (&Paginator{}).new(o)
}

func (p *Paginator) new(o Options) (*Paginator, error) {
	p.options = o
	err := p.decode()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Paginator) Find(tx *gorm.DB, dst interface{}) error {
	if p.options.Model == nil {
		return common.ErrEmptyModelInPaginator
	}
	if p.options.DB == nil {
		return common.ErrEmptyDBInPaginator
	}

	// check what dst is pointer to slice
	if reflect.ValueOf(dst).Kind() != reflect.Ptr {
		return common.ErrInvalidFindDestinationNotPointer
	}
	if reflect.Indirect(reflect.ValueOf(dst)).Kind() != reflect.Slice {
		return common.ErrInvalidFindDestinationNotSlice
	}

	// execute query
	if p.cursor == nil {
		return common.ErrInvalidCursor
	}

	q := p.options.DB
	for k := range tx.Statement.Preloads {
		q = q.Preload(k)
	}

	err := q.Table("(?) as t", tx.Session(&gorm.Session{})).Scopes(p.cursor.Scope()).Find(dst).Error
	// err := tx.Session(&gorm.Session{}).Scopes(p.cursor.Scope()).Find(dst).Error
	if err != nil {
		return err
	}

	// calc paginationinfo
	p.PageInfo = p.calcPageInfo(tx, dst)
	return nil
}

func (p *Paginator) calcPageInfo(tx *gorm.DB, dst interface{}) *PageInfo {
	object := reflect.Indirect(reflect.ValueOf(dst))
	if object.IsNil() || object.Len() == 0 {
		return nil
	}

	// query for totalRow
	totalRows := p.count(tx.Session(&gorm.Session{}))

	// last elem to nextCursor
	nextCursor := p.cursor.ToCursor(object.Index(object.Len() - 1).Interface())

	// first elem to prevCursor
	prevCursor := p.cursor.ToCursor(object.Index(0).Interface())
	// prevCursor.Backward = true

	// save paginationInfo to p
	pageInfo := &PageInfo{
		Next:      nextCursor.Encode(),
		Prev:      prevCursor.SetBackward().Encode(),
		HasNext:   p.checkPage(tx.Session(&gorm.Session{}), nextCursor.Scope()),
		HasPrev:   p.checkPage(tx.Session(&gorm.Session{}), prevCursor.Scope()),
		TotalRows: int(totalRows),
	}

	return pageInfo
}

func (p *Paginator) decode() error {
	if p.options.GinContext == nil {
		return common.ErrEmptyGinContextInPaginator
	}
	sortingQuery := p.options.GinContext.Query("sorting")
	cursorQuery := p.options.GinContext.Query("cursor")

	cursor, err := cursor.DecodeAction(sortingQuery, cursorQuery, p.options.DefaultCursor, p.options.Model, p.options.Limit)
	if err != nil {
		return err
	}

	// cursor.DB = p.DB
	p.cursor = cursor
	return nil
}

func (p *Paginator) count(tx *gorm.DB) (count int64) {
	if err := p.options.DB.Table("(?) as t", tx.Session(&gorm.Session{})).Select("count(1)").Limit(1).Count(&count).Error; err != nil {
		log.Println(err)
		return -1
	}
	return
}

func (p *Paginator) checkPage(tx *gorm.DB, scope func(*gorm.DB) *gorm.DB) (isExist bool) {
	var count int64
	if err := p.options.DB.Table("(?) as t", tx.Session(&gorm.Session{})).Scopes(scope).Select("count(1)").Limit(1).Count(&count).Error; err != nil {
		log.Println(err)
		return
	}
	if count > 0 {
		return true
	}
	return
}
