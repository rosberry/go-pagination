# Pagination library
Package go-pagination was designed to paginate simple RESTful APIs with GORM and Gin.

## Usage

#### Model
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

#### Controllers
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
    "data": {
    },
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


## About

<img src="https://github.com/rosberry/Foundation/blob/master/Assets/full_logo.png?raw=true" height="100" />

This project is owned and maintained by [Rosberry](http://rosberry.com). We build mobile apps for users worldwide 🌏.

Check out our [open source projects](https://github.com/rosberry), read [our blog](https://medium.com/@Rosberry) or give us a high-five on 🐦 [@rosberryapps](http://twitter.com/RosberryApps).

## License

This project is available under the MIT license. See the LICENSE file for more info.
