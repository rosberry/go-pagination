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
			// ID        uint           `gorm:"primary_key"`
			// CreatedAt time.Time      `cursor:"createdAt"`
			// UpdatedAt time.Time      `json:"updated_at"`
			// DeletedAt gorm.DeletedAt `gorm:"index"`

			Name     string
			Role     uint `cursor:"roleID"`
			Clappers []Clapper
		}

		Material struct {
			UserID uint
			User   User `gorm:"foreignKey:UserID"`

			ID        uint           `gorm:"primary_key"`
			CreatedAt time.Time      `cursor:"createdAt"`
			UpdatedAt time.Time      `json:"updated_at"`
			DeletedAt gorm.DeletedAt `gorm:"index"`

			Link    string
			Comment string

			ItemID      string // Unical +
			ItemOwnerID int    // Unical
			ItemType    string `cursor:"item_type_name"`
		}
	)

	type TestDataStruct struct {
		SortName string
		DBName   string
	}

	testData := []TestDataStruct{
		{"user.name", `"User__name"`}, // TODO: Panica!
		{"id", "id"},
		{"comment", "comment"},
		{"item_type_name", "item_type"},
		{"updated_at", "updated_at"},
		{"createdAt", "created_at"},
		{"DeletedAt", "deleted_at"},
	}

	for i, td := range testData {
		name := NSortNameToDBName(td.SortName, &Material{})
		if name != td.DBName {
			t.Errorf("%v) Not equal: %s != %s for %s", i, name, td.DBName, td.SortName)
		}
	}
}
