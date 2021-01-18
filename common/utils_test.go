package common

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestNSortNameToDBName(t *testing.T) {
	type (
		BaseModel struct {
			ID        uint
			CreatedAt time.Time
			UpdatedAt time.Time
			DeletedAt gorm.DeletedAt
		}

		Clapper struct {
			BaseModel
			UserID uint
			Name   string
			Token  string
		}

		User struct {
			BaseModel
			Name     string
			Role     uint `cursor:"roleID"`
			Clappers []Clapper
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
			ItemType    string `cursor:"item_type_name"`

			UserID uint
			User   User `gorm:"foreignKey:UserID"`
		}
	)

	type TestDataStruct struct {
		SortName string
		DBName   string
	}

	var testData = []TestDataStruct{
		{"user.name", `"User".name`},
		{"id", "id"},
		{"comment", "comment"},
		{"item_type_name", "item_type"},
	}

	for i, td := range testData {
		name := NSortNameToDBName(td.SortName, &Material{})
		if name != td.DBName {
			t.Errorf("%v) Not equal: %s != %s for %s", i, name, td.DBName, td.SortName)
		}
	}
}
