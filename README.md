# Pagination library


## Client-Server interaction

Request:
```
GET \items
```

```
GET \items?sorting=%5B%7B%22field%22:%22id%22,%22direction%22:%22desc%22%7D%5D

object in sort:
[
    {
        "field": "Name",
        "direction": "asc"
    },
    {
        "field": "UpdatedAt",
        "direction": "desc"
    }
]

```

Response:
```
{
    "result": true,
    "data": {},
    "pagination": {
        "hasPrev": false,
        "hasNext": true,
        "prev": "ew2YWxU0Cn01ZSI6ogICJID=",
        "next": "ewogICJ2YWx1ZSI6IDU0Cn0="
    }
}
```

Request next page:

```
GET \items?"cursor": "ew2YWxU0Cn01ZSI6ogICJID"
```

Response (end):
```
{
    "result": true,
    "data": null,
    "pagination:" null
}
```

## Basic usage

### Model
```go
import "github.com/rosberry/go-pagination"

//Add ScopeFunc to your database query func
//and use it as .Scopes() argument
func GetItems(count uint, scope pagination.ScopeFunc) []Item {
	var items []Item

	db.DB.Model(&Item{}).Scopes(scope).Where("count < 22").Find(&items)
	return items
}
```

### Controllers
```go
import "github.com/rosberry/go-pagination"

type (
	itemRequest struct {
		Count uint `json:"count" form:"count"`
	}

	responseWithPaging struct {
		Result bool `json:"result"`
		Items  []models.Item
		Paging *pagination2.PaginationResponse
	}
)

func ItemsCursor(c *gin.Context) {
	var request itemRequest
	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, cm.Error[cm.ErrReqInvalid])
		return
	}

    //Decode cursor from request (gin context) 
	cursor, _ := pagination2.Model(&models.Item{}).Decode(c, pagination2.DefaultCursor)

    //Use cursor.Scope() in your model func
	items := models.GetItems(request.Count, cursor.Scope())

    //Use cursor.Result() to get paging response and modify result
	paging, result := cursor.Result(items)

    //Assertion result to your type
	items = result.([]models.Item)

	c.JSON(http.StatusOK, responseWithPaging{
		Result: true,
		Items:  items,
		Paging: paging,
	})
}

```

## Advanced

### Default cursor

#### Base usage
```go
//pagination2.DefaultCursor
cursor, _ := pagination2.Model(&models.Item{}).Decode(c, pagination2.DefaultCursor)
```

#### Base cursor for model
Add embedded struct to your model and you can use DefaultCursorGetter() for model

```go
type Item struct {
    ID         uint
    Title      string

    pagination2.Pagination
}

...
cursor, _ := pagination2.Model(&models.Item{}).Decode(c, item.DefaultCursor)
```

And you can override default getter
```go
func (i *Item) DefaultCursor() *pagination2.Cursor {
	cursor := (&pagination2.Cursor{}).New(5)
	cursor.AddField("name", nil, "desc")

	return cursor
}
```

or create another getter
```go
func (i *Item) AnotherCursor() *pagination2.Cursor {
	cursor := (&pagination2.Cursor{}).New(5)
    cursor.AddField("name", nil, "desc")
    cursor.AddField("title", nil, "asc")

	return cursor
}


func AnotherCursorWithoutModel() *pagination2.Cursor {
	return (&pagination2.Cursor{}).New(5).AddField("name", nil, "desc").AddField("title", nil, "asc")
}

```

and use it
```go
cursor, _ = pagination2.Model(&models.Item{}).Decode(c, item.AnotherCursor)
cursor, _ = pagination2.Model(&models.Item{}).Decode(c, AnotherCursorWithoutModel)
```

### Sorting on client

Using the sorting configuration from client-side is recommended only in exceptional cases. By default, only the immediate fields of the main model support sorting. If you need to sort data by fields from related models ....

As example

```go
type Material struct {
		ID        uint `gorm:"primary_key"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Link    string
		Type    cm.SocialNetworkType //Unical
		Status  Status
		Comment string

		ItemID      string //Unical +
		ItemOwnerID int    //Unical
		ItemType    string

		Claps       int64 `sql:"-"` //calculate from another table
		FailedClaps int64 `sql:"-"` //calculate from another table

		UserID uint
	}


//GetList return all materials (default flow)
//--
//--
func GetList(scope pagination.ScopeFunc) (materials Materials) {
	models.GetDB().Scopes(scope).Find(&materials)
	return
}
func (m *Material) AfterFind(tx *gorm.DB) (err error) {
	m.Claps = claps.SuccessCountByMaterial(m.ID)
	m.FailedClaps = claps.FailedCountByMaterial(m.ID)
	return nil
}


//GetList return all materials (if you want use cursor and sorting)
//--
//--
func GetListForCursor(scope pagination.ScopeFunc) (materials Materials) {
	models.DB.Table("(?) as t", models.DB.
		Table("materials").
		Select(`materials.*, 
			(select count(1) from claps where claps.material_id = materials.id and claps.success = true) as claps,
			(select count(1) from claps where claps.material_id = materials.id and claps.success = false) as failed_claps
			`)).
		Scopes(scope).
		Find(&materials)
	return
}

```