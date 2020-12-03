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
	cursor := pagination2.Decode(c, pagination2.DefaultCursor)

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
cursor := pagination2.Decode(c, pagination2.DefaultCursor)
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
cursor := pagination2.Decode(c, item.DefaultCursor)
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
cursor = pagination2.Decode(c, item.AnotherCursor)
cursor = pagination2.Decode(c, AnotherCursorWithoutModel)
```

