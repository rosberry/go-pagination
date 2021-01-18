package controllers

import (
	"fmt"
	"log"

	"github.com/rosberry/go-pagination/_examples/simple/models"

	"github.com/rosberry/go-pagination"

	"github.com/gin-gonic/gin"
)

type (
	userData struct {
		ID       uint             `json:"id"`
		Name     string           `json:"name"`
		Role     uint             `json:"roleID"`
		RoleName string           `json:"role"`
		Clappers []models.Clapper `json:"clappers"`
	}

	usersListResponse struct {
		Result     bool
		Users      []userData
		Pagination *pagination.PageInfo
	}

	materialsListResponse struct {
		Result     bool
		Materials  []models.Material
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
			Clappers: um.Clappers,
		}
	}

	return ud
}

func UsersList(c *gin.Context) {
	paginator, err := pagination.New(pagination.Options{
		GinContext: c,
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

func Materials(c *gin.Context) {
	paginator, err := pagination.New(pagination.Options{
		GinContext: c,
	})
	if err != nil {
		log.Println(err)
	}

	m := models.GetMaterialsList(paginator)

	c.JSON(200, materialsListResponse{
		Result:     true,
		Materials:  m,
		Pagination: paginator.PageInfo,
	})
}
