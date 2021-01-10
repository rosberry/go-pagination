package common

import "errors"

var (
	ErrInvalidCursor                    = errors.New("Invalid cursor")
	ErrInvalidSorting                   = errors.New("Invalid sorting")
	ErrInvalidDefaultCursor             = errors.New("Invalid default cursor")
	ErrCursorAndSortingTogether         = errors.New("You cannot use cursor and sorting at the same time")
	ErrInvalidFindDestinationNotPointer = errors.New("dst is not pointer to slice")
	ErrInvalidFindDestinationNotSlice   = errors.New("pointer in dst not to slice")
)