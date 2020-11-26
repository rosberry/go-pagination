# Pagination library


# Client-Server interaction

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