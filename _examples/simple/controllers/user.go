package controllers

import (
	"fmt"
	"log"
	"pagination"
	"pagination/_examples/simple/models"

	"github.com/gin-gonic/gin"
)

type (
	userData struct {
		ID       uint   `json:"id"`
		Name     string `json:"name"`
		Role     uint   `json:"roleID"`
		RoleName string `json:"role"`
	}

	usersListResponse struct {
		Result     bool
		Users      []userData
		Pagination *pagination.PageInfo
	}
)

func usersListToData(u []models.User) []userData {
	ud := make([]userData, len(u), len(u))
	for i, um := range u {
		ud[i] = userData{
			ID:       um.ID,
			Name:     um.Name,
			Role:     um.Role,
			RoleName: fmt.Sprintf("stub %v", i),
		}
	}

	return ud
}

func UsersList(c *gin.Context) {
	paginator, err := pagination.New(pagination.Options{
		GinContext: c,
		Limit:      2,
		Model:      &models.User{},
	})
	if err != nil {
		log.Println(err)
	}

	log.Printf("Paginator: %+v\n", paginator)
	users := models.GetUsersList(0, paginator)

	c.JSON(200, usersListResponse{
		Result:     true,
		Users:      usersListToData(users),
		Pagination: paginator.PageInfo,
	})
}
