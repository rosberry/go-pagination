package main

import (
	"log"
	"pagination/_examples/simple/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("It's simple example")

	//pagination.New(pagination.Options{})
	s := setupServer()
	s.Run()
}

func setupServer() *gin.Engine {
	r := gin.Default()

	r.GET("/list", controllers.UsersList)

	return r
}