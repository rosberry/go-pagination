package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type (
	//DirectionType ...
	DirectionType string

	//Cursor struct
	Cursor struct {
		Fields   []Field `json:"fields"`
		Limit    int     `json:"limit"`
		Backward bool    `json:"backward"`
	}

	//Field struct
	Field struct {
		Name      string        `json:"name"`
		Value     interface{}   `json:"value"`
		Direction DirectionType `json:"direction"`
	}
)

//New Cursor
func (c *Cursor) New(limit int, fields ...Field) *Cursor {
	return &Cursor{
		Fields: fields,
		Limit:  limit,
	}
}

//AddField to cursor
func (c *Cursor) AddField(name string, value interface{}, order DirectionType) *Cursor {
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
				s := fmt.Sprintf("%v %v ?", c.Fields[j].Name, CompareTerms[c.Fields[j].Direction.Backward(c.Backward)])
				val = append(val, c.Fields[j].Value)
				if j != 0 {
					query += " AND "
				}
				query += s
			}
		}
		//query += ")"
		// log.Println("Additional query:", query)
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
		// log.Println("order:", order)
		query = query.Order(order)
		if c.Limit != 0 {
			query = query.Limit(c.Limit + 1)
		}
	}

	return query
}

//GroupConditions for GORM v2
func (c *Cursor) GroupConditions(db *gorm.DB) *gorm.DB {
	return c.order(c.where(db))
}

//Backward of order type
func (dt DirectionType) Backward(ok bool) DirectionType {
	if !ok {
		return dt
	}
	switch dt {
	case DirectionDesc:
		return DirectionAsc
	case DirectionAsc:
		return DirectionDesc
	default:
		return DirectionAsc
	}
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

//Result - create new cursors by result list and modify items
func (c *Cursor) Result(items interface{}) (*PaginationResponse, interface{}) {
	log.Printf("Init cursor: %+v\n", c)

	if reflect.TypeOf(items).Kind() != reflect.Slice {
		return nil, nil
	}

	object := reflect.ValueOf(items)

	if object.Len() == 0 {
		return nil, nil
	}

	nextCursor := (&Cursor{}).New(c.Limit)
	prevCursor := (&Cursor{}).New(c.Limit)
	prevCursor.Backward = true

	var hasNext, hasPrev bool

	hasNext = object.Len() > c.Limit
	if hasNext {
		object = object.Slice(0, c.Limit)
	}
	if c.Backward {
		object = revert(object)
	}

	first := object.Index(0)
	last := object.Index(object.Len() - 1)

	typ := first.Type()

	var fieldSearch func(f Field, typ reflect.Type, indxs ...int)
	fieldSearch = func(f Field, typ reflect.Type, indxs ...int) {
		for i := 0; i < typ.NumField(); i++ {
			structField := typ.Field(i)
			log.Println(i, structField.Name)

			if structField.Type.Kind() == reflect.Struct && structField.Anonymous {
				fieldSearch(f, structField.Type, append(indxs, i)...)
			}

			log.Println("structField.Name:", structField.Name)
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

			log.Println(f.Name, ":", name)
			if f.Name == name {
				nextCursor.AddField(name, lastVal, f.Direction)
				prevCursor.AddField(name, firstVal, f.Direction)
			}

			if f.Value != nil {
				hasPrev = true
			}

		}
	}

	//log.Println("name, val:", name, firstVal, lastVal)
	for _, f := range c.Fields {
		fieldSearch(f, typ)
		continue
	}

	//items = object.Interface() //??--

	log.Printf("Prev cursor: %+v\n", prevCursor)
	log.Printf("Next cursor: %+v\n", nextCursor)
	log.Println("Has next:", hasNext)
	log.Println("Has prev:", hasPrev)

	return &PaginationResponse{
		Next:    nextCursor.Encode(),
		Prev:    prevCursor.Encode(),
		HasNext: hasNext,
		HasPrev: hasPrev,
	}, object.Interface()
}

func columnName(field reflect.StructField) string {
	tags := field.Tag

	var colName string

	colName = tags.Get("cursor")
	if colName == "" {
		colName = (&schema.NamingStrategy{}).ColumnName("", field.Name)
	}
	return colName
}

func revert(object reflect.Value) reflect.Value {
	if object.Len() <= 1 {
		return object
	}
	result := reflect.MakeSlice(object.Type(), object.Len(), object.Cap())
	for i := 0; i < object.Len(); i++ {
		result.Index(i).Set(object.Index(object.Len() - 1 - i))
	}
	return result
}

func sortNameToDBName(sortName string, typ reflect.Type) string {
	for i := 0; i < typ.NumField(); i++ {
		structField := typ.Field(i)
		name := columnName(structField)

		if sortName == name {
			field := (&schema.Schema{}).ParseField(structField)
			if field.DBName != "" {
				return field.DBName
			}
			return (&schema.NamingStrategy{}).ColumnName("", field.Name)
		}
	}

	return sortName
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
