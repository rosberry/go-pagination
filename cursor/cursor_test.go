package cursor

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

func TestFindFieldValueByFieldName(t *testing.T) {
	type (
		BaseModel struct {
			ID        uint `gorm:"primary_key"`
			CreatedAt time.Time
			UpdatedAt time.Time
		}

		BaseModelWithSoftDelete struct {
			BaseModel
			DeletedAt gorm.DeletedAt `gorm:"index"`
		}

		//User is the user model of the mobile application.
		User struct {
			BaseModelWithSoftDelete
			Name  string
			Photo string
		}

		//Users list
		Users []User

		Status uint

		Material struct {
			ID        uint `gorm:"primary_key"`
			CreatedAt time.Time
			UpdatedAt time.Time
			DeletedAt gorm.DeletedAt `gorm:"index"`

			Link    string
			Status  Status
			Comment string

			ItemID      string `cursor:"item_id_cursor"` //Unical +
			ItemOwnerID int    //Unical
			ItemType    string

			Claps       int64 `sql:"-" gorm:"-"` //calculate
			FailedClaps int64 `sql:"-" gorm:"-"` //calc

			UserID     uint
			Author     User `gorm:"foreignKey:UserID"`
			LikesCount uint
		}

		Materials []Material
	)

	material := Material{
		ID:      5,
		Comment: "test material",
		Author:  User{BaseModelWithSoftDelete{BaseModel: BaseModel{ID: 0}}, "Ivan", ""},
	}

	/*
		v := searchFieldValue("comment", &material)
		if v != material.Comment {
			t.Errorf("%v", v)
		}
	*/
	v := searchFieldValue(`"Author__name"`, &material)
	if v != material.Author.Name {
		t.Errorf("%v", v)
	}
}
