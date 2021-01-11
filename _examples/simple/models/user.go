package models

import (
	"log"
	"pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID   uint
	Name string
	Role uint `cursor:"roleID"`
}

func GetUsersList(role uint, paging *pagination.Paginator) []User {
	//db := mockDB().Session(&gorm.Session{DryRun: true})
	db := liveDB().Session(&gorm.Session{DryRun: false})

	var users []User
	q := db.Model(&User{}).Where("role = ?", role)

	err := paging.Find(q, &users)
	if err != nil {
		log.Println(err)
		return nil
	}

	return users
}

//DB connection
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
	connString := "host=localhost port=5432 user=postgres dbname=clapper password=123 sslmode=disable"
	db, err := gorm.Open(postgres.Open(connString), gormConf)
	if err != nil {
		log.Println(err)
		return nil
	}
	return db.Debug()
}
