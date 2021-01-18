package cursor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/rosberry/go-pagination/common"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type (
	//Cursor struct
	Cursor struct {
		Fields   []Field `json:"fields"`
		Limit    int     `json:"limit"`
		Backward bool    `json:"backward"`

		DB *gorm.DB `json:"-"`
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
	var qList string
	val := make([]interface{}, 0)
	//Make cursor query
	for i, f := range c.Fields {
		if f.Value == nil {
			continue
		}
		//query := "("
		query := ""
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

		qList += fmt.Sprintf(" OR (%v)", query)
	}
	qList = strings.Replace(qList, " OR (", "(", 1)
	q = q.Where(qList, val...)

	//-------
	return q
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
	cursor.DB = c.DB
	typ := value.Type()

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

			if f.Name == name {
				cursor.AddField(name, val, f.Direction)
			}
		}
	}

	for _, f := range c.Fields {
		fieldSearch(f, typ)
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
