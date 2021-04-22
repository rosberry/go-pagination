package cursor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/rosberry/go-pagination/common"
)

type (
	// Cursor struct
	Cursor struct {
		Fields   []Field `json:"fields"`
		Limit    int     `json:"limit"`
		Backward bool    `json:"backward"`

		DB *gorm.DB `json:"-"`
	}

	// Field struct
	Field struct {
		Name      string               `json:"name"`
		Value     interface{}          `json:"value"`
		Direction common.DirectionType `json:"direction"`
	}
)

// New Cursor
func New(limit int, fields ...Field) *Cursor {
	return &Cursor{
		Fields: fields,
		Limit:  limit,
	}
}

func (c *Cursor) SetBackward() *Cursor {
	if c == nil {
		return nil
	}

	c.Backward = true

	return c
}

// AddField to cursor
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

// Scope convert Cursor to gorm.DB query
func (c *Cursor) Scope() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return c.order(c.where(db))
	}
}

// where convertation
func (c *Cursor) where(db *gorm.DB) *gorm.DB {
	q := db

	var qList string

	val := make([]interface{}, 0)
	// Make cursor query
	for i, f := range c.Fields {
		if f.Value == nil {
			continue
		}
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

// order convertation
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

// GroupConditions for GORM v2
func (c *Cursor) GroupConditions(db *gorm.DB) *gorm.DB {
	return c.order(c.where(db))
}

// Encode Cursor to base64 string
func (c *Cursor) Encode() string {
	raw, err := json.Marshal(c)
	if err != nil {
		log.Println("Marshal err:", err)
		return ""
	}

	base64str := base64.StdEncoding.EncodeToString(raw)

	return base64str
}

func (c *Cursor) ToCursor(value interface{}) (cursor *Cursor) {
	cursor = New(c.Limit)
	cursor.DB = c.DB

	for _, f := range c.Fields { // f.Name = `"Author__name"`
		val := searchFieldValue(f.Name, value)
		if val != nil {
			cursor.AddField(f.Name, val, f.Direction)
		} else {
			log.Print("!!!")
		}
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

func searchFieldValue(name string, in interface{}) (value interface{}) {
	// modify sortName
	namesChain := strings.Split(strings.Trim(name, `"`), "__")

	for _, n := range namesChain {
		end, value := findFieldValueByFieldName(n, in)
		if end {
			return value
		}

		in = value
	}

	return nil
}

func findFieldValueByFieldName(name string, model interface{}) (end bool, value interface{}) {
	name = strings.ToLower(name)

	var (
		typ reflect.Type
		val reflect.Value
	)

	if reflect.ValueOf(model).Kind() == reflect.Ptr {
		typ = reflect.Indirect(reflect.ValueOf(model)).Type()
		val = reflect.Indirect(reflect.ValueOf(model))
	} else {
		typ = reflect.ValueOf(model).Type()
		val = reflect.ValueOf(model)
	}

	if typ.Kind() != reflect.Struct {
		return true, model
	}

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Type.Kind() == reflect.Struct && f.Anonymous {
			end, v := findFieldValueByFieldName(name, val.Field(i).Interface())
			if end {
				return end, v
			}
		}

		if fName := fieldNameByDBName(f); name == fName {
			fTyp := f.Type

			end = true
			if fTyp.Kind() == reflect.Struct {
				switch {
				case fTyp.AssignableTo(reflect.TypeOf(time.Time{})):
				case fTyp.AssignableTo(reflect.TypeOf(gorm.DeletedAt{})):
				default:
					end = false
				}
			}
			return end, val.Field(i).Interface()
		}
	}

	return false, model
}
