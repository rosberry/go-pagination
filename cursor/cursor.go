package cursor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"pagination/common"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type (
	//Cursor struct
	Cursor struct {
		Fields   []Field `json:"fields"`
		Limit    int     `json:"limit"`
		Backward bool    `json:"backward"`

		db *gorm.DB
	}

	//Field struct
	Field struct {
		Name      string               `json:"name"`
		Value     interface{}          `json:"value"`
		Direction common.DirectionType `json:"direction"`
	}
)

//New Cursor
func New(limit int, fields ...Field) *Cursor {
	return &Cursor{
		Fields: fields,
		Limit:  limit,
	}
}

//AddField to cursor
func (c *Cursor) AddField(name string, value interface{}, order common.DirectionType) *Cursor {
	if c == nil {
		return nil
	}
	for _, f := range c.Fields {
		if f.Name == name {
			f.Value = value
			f.Direction = order
			return c
		}
	}

	c.Fields = append(c.Fields, Field{
		Name:      name,
		Value:     value,
		Direction: order,
	})

	return c
}

//Scope convert Cursor to gorm.DB query
func (c *Cursor) Scope() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return c.order(c.where(db))
	}
}

//where convertation
func (c *Cursor) where(db *gorm.DB) *gorm.DB {
	q := db
	//Make cursor query
	for i, f := range c.Fields {
		if f.Value == nil {
			continue
		}
		//query := "("
		query := ""
		val := make([]interface{}, 0)
		for j := 0; j <= i; j++ {
			if j != i {
				s := fmt.Sprintf("%v %v ?", c.Fields[j].Name, "=")
				val = append(val, c.Fields[j].Value)
				if j != 0 {
					query += " AND "
				}
				query += s
			} else {
				s := fmt.Sprintf("%v %v ?", c.Fields[j].Name, common.CompareTerms[c.Fields[j].Direction.Backward(c.Backward)])
				val = append(val, c.Fields[j].Value)
				if j != 0 {
					query += " AND "
				}
				query += s
			}
		}

		if len(c.Fields) != 1 {
			q = q.Or(query, val...)
		} else {
			q = q.Where(query, val...)
		}
	}
	//-------
	return db.Where(q)
}

//order convertation
func (c *Cursor) order(query *gorm.DB) *gorm.DB {
	for _, f := range c.Fields {
		order := fmt.Sprintf("%s %s", f.Name, f.Direction.Backward(c.Backward))

		query = query.Order(order)
		if c.Limit != 0 {
			query = query.Limit(c.Limit)
		}
	}

	return query
}

//GroupConditions for GORM v2
func (c *Cursor) GroupConditions(db *gorm.DB) *gorm.DB {
	return c.order(c.where(db))
}

//Encode Cursor to base64 string
func (c *Cursor) Encode() string {
	raw, err := json.Marshal(c)
	if err != nil {
		log.Println("Marshal err:", err)
		return ""
	}

	base64str := base64.StdEncoding.EncodeToString(raw)
	return base64str
}

func (c *Cursor) ToCursor(value reflect.Value) (cursor *Cursor) {
	cursor = New(common.DefaultLimit)
	typ := value.Type()
	//value := reflect.ValueOf(elem)
	log.Println(value)
	var fieldSearch func(f Field, typ reflect.Type, indxs ...int)
	fieldSearch = func(f Field, typ reflect.Type, indxs ...int) {
		for i := 0; i < typ.NumField(); i++ {
			structField := typ.Field(i)
			if structField.Type.Kind() == reflect.Struct && structField.Anonymous {
				fieldSearch(f, structField.Type, append(indxs, i)...)
			}
			name := fieldNameByDBName(structField)

			var val interface{}
			var v reflect.Value

			if len(indxs) > 0 {
				for i, indx := range indxs {
					if i == 0 {
						v = value.Field(indx)
					} else {
						v = v.Field(indx)
					}

				}
				val = v.Field(i).Interface()
			} else {
				val = value.Field(i).Interface()
			}

			log.Println("f.Name == name", f.Name, name)
			if f.Name == name {
				cursor.AddField(name, val, f.Direction)
				log.Println("Add field:", cursor)
			}
		}
	}

	for _, f := range c.Fields {
		fieldSearch(f, typ)
		continue
	}
	return
}

func fieldNameByDBName(f reflect.StructField) string {
	if f.Anonymous {
		return f.Name
	}
	field := (&schema.Schema{}).ParseField(f)
	if field.DBName != "" {
		return field.DBName
	}
	return (&schema.NamingStrategy{}).ColumnName("", f.Name)
}

//Result - create new cursors by result list and modify items
/*
func (c *Cursor) Result(items interface{}) (*PaginationResponse, interface{}) {
	log.Printf("Init cursor: %+v\n", c)

	if reflect.TypeOf(items).Kind() != reflect.Slice {
		return nil, nil
	}

	object := reflect.ValueOf(items)

	if object.Len() == 0 {
		return nil, nil
	}

	nextCursor := New(c.Limit)
	prevCursor := New(c.Limit)
	prevCursor.Backward = true

	var hasNext, hasPrev bool

	if c.Backward {
		hasPrev = object.Len() > c.Limit
		if hasPrev {
			object = object.Slice(0, c.Limit)
		}
		hasNext = true
		object = revert(object)
	} else {
		hasNext = object.Len() > c.Limit
		if hasNext {
			object = object.Slice(0, c.Limit)
		}
		if len(c.Fields) > 0 && c.Fields[0].Value != nil {
			hasPrev = true
		}
	}

	first := object.Index(0)
	last := object.Index(object.Len() - 1)

	typ := first.Type()

	var fieldSearch func(f Field, typ reflect.Type, indxs ...int)
	fieldSearch = func(f Field, typ reflect.Type, indxs ...int) {
		for i := 0; i < typ.NumField(); i++ {
			structField := typ.Field(i)
			if structField.Type.Kind() == reflect.Struct && structField.Anonymous {
				fieldSearch(f, structField.Type, append(indxs, i)...)
			}
			name := fieldNameByDBName(structField)

			var firstVal, lastVal interface{}
			var fv, lv reflect.Value

			if len(indxs) > 0 {
				for i, indx := range indxs {
					if i == 0 {
						fv, lv = first.Field(indx), last.Field(indx)
					} else {
						fv, lv = fv.Field(indx), lv.Field(indx)
					}

				}
				firstVal, lastVal = fv.Field(i).Interface(), lv.Field(i).Interface()
			} else {
				firstVal, lastVal = first.Field(i).Interface(), last.Field(i).Interface()
			}

			if f.Name == name {
				nextCursor.AddField(name, lastVal, f.Direction)
				prevCursor.AddField(name, firstVal, f.Direction)
			}

			if f.Value != nil {
				hasPrev = true
			}

		}
	}

	for _, f := range c.Fields {
		fieldSearch(f, typ)
		continue
	}

	log.Printf("Prev cursor: %+v\n", prevCursor)
	log.Printf("Next cursor: %+v\n", nextCursor)
	log.Println("Has next:", hasNext)
	log.Println("Has prev:", hasPrev)

	resp := &PaginationResponse{
		Next:    nextCursor.Encode(),
		Prev:    prevCursor.Encode(),
		HasNext: hasNext,
		HasPrev: hasPrev,
	}

	if c.db != nil {
		var count int64
		if err := c.db.Table(tableName(typ)).Count(&count).Error; err == nil {
			resp.TotalRows = int(count)
		}

	}

	return resp, object.Interface()
}
*/