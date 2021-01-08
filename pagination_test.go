package pagination

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
	//model
	db := mockDB()
	db = db.Session(&gorm.Session{DryRun: true})

	type User struct {
		ID    uint
		Name  string
		Count uint
	}
	var user User

	stmt := db.Find(user).Statement
	sql := stmt.SQL.String()

	log.Println(sql)

	//controller

}
