package pagination

import "errors"

const (
	DirectionDesc DirectionType = "desc"
	DirectionAsc  DirectionType = "asc"
)

var CompareTerms map[DirectionType]string = map[DirectionType]string{
	DirectionDesc: "<",
	DirectionAsc:  ">",
}

var DirectionByString map[string]DirectionType = map[string]DirectionType{
	"asc":  DirectionAsc,
	"desc": DirectionDesc,
}

const (
	defaultLimit = 3
)

var (
	ErrInvalidCursor            = errors.New("Invalid cursor")
	ErrInvalidSorting           = errors.New("Invalid sorting")
	ErrInvalidDefaultCursor     = errors.New("Invalid default cursor")
	ErrCursorAndSortingTogether = errors.New("You cannot use cursor and sorting at the same time")
)
