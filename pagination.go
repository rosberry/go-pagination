package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"paging/db"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type (
	//OrderType ...
	OrderType string

	//Cursor struct
	Cursor struct {
		Fields []Field
		Limit  int
	}

	//Field struct
	Field struct {
		Name  string
		Value interface{}
		Order OrderType
	}

	//Pagination struct
	Pagination struct {
		Next    string `json:"next"`
		Prev    string `json:"prev"`
		HasNext bool   `json:"hasNext"`
		HasPrev bool   `json:"hasPrev"`
	}
)

const (
	OrderDesc OrderType = "desc"
	OrderAsc  OrderType = "asc"
)

var CompareTerms map[OrderType]string = map[OrderType]string{
	OrderDesc: "<",
	OrderAsc:  ">",
}

//Backward of order type
func (o OrderType) Backward() OrderType {
	switch o {
	case OrderDesc:
		return OrderAsc
	case OrderAsc:
		return OrderDesc
	default:
		return OrderAsc
	}
}

//New Cursor
func (c *Cursor) New(limit int, fields ...Field) *Cursor {
	return &Cursor{
		Fields: fields,
		Limit:  limit,
	}
}

//AddField to cursor
func (c *Cursor) AddField(name string, value interface{}, order OrderType) *Cursor {
	for _, f := range c.Fields {
		if f.Name == name {
			f.Value = value
			f.Order = order
			return c
		}
	}

	c.Fields = append(c.Fields, Field{
		Name:  name,
		Value: value,
		Order: order,
	})

	return c
}

//where convertation
func (c *Cursor) where() *gorm.DB {
	q := db.DB
	//Make cursor query
	for i, f := range c.Fields {
		if f.Value == nil {
			continue
		}
		query := "("
		val := make([]interface{}, 0)
		for j := 0; j <= i; j++ {
			if j != i {
				s := fmt.Sprintf("(%v %v ?)", c.Fields[j].Name, "=")
				val = append(val, c.Fields[j].Value)
				if j != 0 {
					query += " AND "
				}
				query += s
			} else {
				s := fmt.Sprintf("(%v %v ?)", c.Fields[j].Name, CompareTerms[c.Fields[j].Order])
				val = append(val, c.Fields[j].Value)
				if j != 0 {
					query += " AND "
				}
				query += s
			}
		}
		query += ")"
		log.Println("Additional query:", query)
		q = q.Or(query, val...)
	}
	//-------
	return q
}

//order convertation
func (c *Cursor) order(query *gorm.DB) *gorm.DB {
	for _, f := range c.Fields {
		order := fmt.Sprintf("%s %s", f.Name, f.Order)
		query = query.Order(order)
		if c.Limit != 0 {
			query = query.Limit(c.Limit + 1)
		}
	}

	return query
}

//Query modify with cursor
func (c *Cursor) Query(query *gorm.DB) *gorm.DB {
	query = c.order(query.Where(c.where()))
	return query
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

//DecodeCursor from base64 string
func DecodeCursor(s string) *Cursor {
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

//Pagination by cursor from items
func (c *Cursor) Pagination(items interface{}) *Pagination {
	log.Printf("Init cursor: %+v\n", c)

	object := reflect.ValueOf(items)
	if object.Type().Kind() != reflect.Slice {
		return nil
	}

	nextCursor := (&Cursor{}).New(c.Limit)
	prevCursor := (&Cursor{}).New(c.Limit)

	var hasNext, hasPrev bool

	hasNext = object.Len() > c.Limit

	if object.Len() > 0 {
		first := object.Index(0)
		last := object.Index(c.Limit - 1)

		typ := first.Type()

		//log.Println("name, val:", name, firstVal, lastVal)
		for _, f := range c.Fields {
			for i := 0; i < typ.NumField(); i++ {
				name := columnName(typ.Field(i))
				firstVal := first.Field(i).Interface()
				lastVal := last.Field(i).Interface()
				if f.Name == name {
					nextCursor.AddField(name, lastVal, f.Order)
					prevCursor.AddField(name, firstVal, f.Order.Backward())
				}

				if f.Value != nil {
					hasPrev = true
				}
			}
		}

	}

	log.Printf("Prev cursor: %+v\n", prevCursor)
	log.Printf("Next cursor: %+v\n", nextCursor)
	log.Println("Has next:", hasNext)
	log.Println("Has prev:", hasPrev)

	return &Pagination{
		Next:    nextCursor.Encode(),
		Prev:    prevCursor.Encode(),
		HasNext: hasNext,
		HasPrev: hasPrev,
	}
}

func columnName(field reflect.StructField) string {
	return (&schema.NamingStrategy{}).ColumnName("", field.Name)
}
