package main

type Product struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Image      *string  `json:"image"`
	Price      float64  `json:"price"`
	Qty        int      `json:"qty"`
	OutOfStock bool     `json:"out_of_stock"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}