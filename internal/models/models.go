package models

type InvoiceData struct {
	CustomerName string  `json:"customer_name"`
	Items        []Item  `json:"items"`
	Total        float64 `json:"total"`
}

type Item struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}
