package src

import (
	de "github.com/shopspring/decimal"
	"os"
)

type (

	// Car
	// @String
	// @Table
	Car struct {
		Name  string                 `json:"name" gorm:"default:'';not null"`
		Price de.Decimal             `json:"price"`
		File  *os.File               `json:"file"`
		Title map[string]interface{} `json:"title"`
		Age   Int                    `json:"age"`
		BT    []int                  `json:"bt"`
	}
)

// This comment is associated with the hello constant.
const hello = "Hello, World!" // line comment 1

// This is my int type
type Int int

// This comment is associated with the foo variable.
var foo = "hello" // line comment 2
