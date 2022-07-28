package src

import (
	de "github.com/shopspring/decimal"
	"github.com/wule61/macro"
	"os"
)

type Car struct {
	macro.Annotator
	Name  string                 `json:"name" gorm:"default:'';not null" macro:""`
	Price de.Decimal             `json:"price"`
	File  *os.File               `json:"file"`
	Title map[string]interface{} `json:"title"`
	BT    []int                  `json:"bt"`
}

func (Car) Annotations() []string {
	return []string{"1", "2"}
}
