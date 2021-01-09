package cursor

import "gorm.io/gorm"

type (
	//DirectionType ...
	DirectionType string

	//Cursor struct
	Cursor struct {
		Fields   []Field `json:"fields"`
		Limit    int     `json:"limit"`
		Backward bool    `json:"backward"`

		db *gorm.DB
	}

	//Field struct
	Field struct {
		Name      string        `json:"name"`
		Value     interface{}   `json:"value"`
		Direction DirectionType `json:"direction"`
	}
)
