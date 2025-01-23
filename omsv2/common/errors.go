package common

import "errors"


//Error message created to be used 
var (
	ErrNoItems = errors.New("items must have at least one item")
	ErrNoStock = errors.New("some item is not in stock")
)