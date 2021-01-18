package models

import (
	"log"
	"time"

	"github.com/rosberry/go-pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type (
	BaseModel struct {
		ID        uint
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt
	}

	User struct {
		BaseModel
		Name     string
		Role     uint `cursor:"roleID"`
		Clappers []Clapper
	}

	Clapper struct {
		BaseModel
		UserID uint
		Name   string
		Token  string
	}

	Material struct {
		ID        uint `gorm:"primary_key"`
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt `gorm:"index"`

		Link    string
		Comment string

		ItemID      string //Unical +
		ItemOwnerID int    //Unical
		ItemType    string

		UserID uint
		User   User `gorm:"foreignKey:UserID"`
	}
)

func GetUsersList(role uint, paginator *pagination.Paginator) []User {
	//db := mockDB().Session(&gorm.Session{DryRun: true})
	db := liveDB().Session(&gorm.Session{DryRun: false})

	paginator.Options.Limit = 2
	paginator.Options.DB = db
	paginator.Options.Model = &User{}

	var users []User
	q := db.Model(&User{}).Preload("Clappers").Where("id < ?", 190)

	err := paginator.Find(q, &users)
	if err != nil {
		log.Println(err)
		return nil
	}

	log.Printf("User: %+v", users)
	return users
}

//GetMaterialsList return all materials
func GetMaterialsList(paginator *pagination.Paginator) (materials []Material) {
	db := liveDB().Session(&gorm.Session{DryRun: false})

	paginator.Options.Limit = 2
	paginator.Options.DB = db
	paginator.Options.Model = &Material{}

	q := db.Model(&Material{}).Joins("User")

	paginator.Find(q, &materials)
	return
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
