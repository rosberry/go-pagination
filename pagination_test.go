package pagination

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var gormConf = &gorm.Config{
	PrepareStmt: true,
}

func mockDB() *gorm.DB {
	sqlDB, _, _ := sqlmock.New()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), gormConf)
	if err != nil {
		log.Println(err)
		return nil
	}
	return db.Debug()
}

func liveDB() *gorm.DB {
	connString := ""
	db, err := gorm.Open(postgres.Open(connString), gormConf)
	if err != nil {
		log.Println(err)
		return nil
	}
	return db.Debug()
}

func TestMainFlow(t *testing.T) {
	//http request-response
	ts := httptest.NewServer(setupServer())
	defer ts.Close()

	// Make a request to our server with the {base url}/ping
	resp, err := http.Get(fmt.Sprintf("%s/list", ts.URL))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %v", resp.StatusCode)
	}

}

func setupServer() *gin.Engine {
	r := gin.Default()

	r.GET("/list", usersList)

	return r
}

//controller
type (
	userData struct {
		ID       uint
		Name     string
		Role     uint
		RoleName string
	}

	usersListResponse struct {
		Result bool
		Users  []userData
	}
)

func usersList(c *gin.Context) {
	getUsersList(0)

	c.JSON(200, gin.H{
		"message": "pong",
	})
}

//model
type User struct {
	ID   uint
	Name string
	Role uint
}

func getUsersList(role uint) []User {
	db := mockDB().Session(&gorm.Session{DryRun: true})

	var users []User
	q := db.Where("role = ?", role).Find(users)

	stmt := q.Statement
	sql := stmt.SQL.String()
	log.Println(sql)

	return nil
}
