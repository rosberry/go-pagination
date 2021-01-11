# Pagination library


## Client-Server interaction

Request:
```
GET /items
```

```
GET /items?sorting=%5B%7B%22field%22:%22id%22,%22direction%22:%22desc%22%7D%5D

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
        "next": "ewogICJ2YWx1ZSI6IDU0Cn0=",
		"totalRows": 10
    }
}
```

Request next page:

```
GET /items?cursor=ew2YWxU0Cn01ZSI6ogICJID
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
package models

import "github.com/rosberry/go-pagination"


type User struct {
	ID   uint
	Name string
	Role uint `json:"roleID" cursor:"roleID"`
}

func GetUsersList(role uint, paginator *pagination.Paginator) []User {
	var users []User
	q := db.DB.Model(&User{}).Where("role = ?", role)

	err := paginator.Find(q, &users)
	if err != nil {
		log.Println(err)
		return nil
	}

	return users
}
```

### Controllers
```go
import "github.com/rosberry/go-pagination"

type (
	usersListResponse struct {
		Result     bool
		Users      []userData
		Pagination *pagination.PageInfo
	}
)


func UsersList(c *gin.Context) {
	paginator, err := pagination.New(pagination.Options{
		GinContext: c,
		//next can be move to model as paginator.Model = &User{} etc...
		DB: models.GetDB(),
		Model: &models.User{},
		Limit: 5,
		DefaultCursor: nil,
	})
	if err != nil {
		log.Println(err)
	}

	users := models.GetUsersList(0, paginator)

	c.JSON(200, usersListResponse{
		Result:     true,
		Users:      usersListToData(users),
		Pagination: paginator.PageInfo,
	})
}

```