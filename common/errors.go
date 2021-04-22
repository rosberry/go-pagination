package common

import "errors"

var (
	ErrInvalidCursor                    = errors.New("invalid cursor")
	ErrInvalidSorting                   = errors.New("invalid sorting")
	ErrInvalidDefaultCursor             = errors.New("invalid default cursor")
	ErrCursorAndSortingTogether         = errors.New("you cannot use cursor and sorting at the same time")
	ErrInvalidFindDestinationNotPointer = errors.New("dst is not pointer to slice")
	ErrInvalidFindDestinationNotSlice   = errors.New("pointer in dst not to slice")
	ErrEmptyModelInPaginator            = errors.New("paginator.Model is nil")
	ErrEmptyDBInPaginator               = errors.New("paginator.DB is nil")
	ErrEmptyGinContextInPaginator       = errors.New("paginator.GinContext is nil")
)
