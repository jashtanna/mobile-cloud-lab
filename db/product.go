package db

import "time"

type Product struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Image      *string   `json:"image,omitempty"`
	Price      float64   `json:"price"`
	Qty        int       `json:"qty"`
	OutOfStock bool      `json:"out_of_stock"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
