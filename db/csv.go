package db

import (
    "encoding/csv"
    "io"
    "log"
    "strconv"
    "strings"
)

func ParseCSV(reader io.Reader) ([]Product, error) {
    csvReader := csv.NewReader(reader)
    csvReader.TrimLeadingSpace = true
    records, err := csvReader.ReadAll()
    if err != nil {
        log.Printf("Failed to read CSV: %v", err)
        return nil, err
    }

    var products []Product
    for i, row := range records[1:] { // Skip header
        if len(row) < 4 {
            log.Printf("Invalid row at line %d: %v", i+2, row)
            continue
        }
        name := strings.TrimSpace(row[0])
        if name == "" {
            log.Printf("Empty name at line %d", i+2)
            continue
        }
        image := strings.TrimSpace(row[1])
        price, err := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
        if err != nil || price < 0 {
            log.Printf("Invalid price at line %d: %s", i+2, row[2])
            continue
        }
        qty, err := strconv.Atoi(strings.TrimSpace(row[3]))
        if err != nil || qty < 0 {
            log.Printf("Invalid quantity at line %d: %s", i+2, row[3])
            continue
        }
        product := Product{
            Name:       name,
            Image:      &image,
            Price:      price,
            Qty:        qty,
            OutOfStock: qty == 0,
        }
        if image == "" {
            product.Image = nil
        }
        products = append(products, product)
    }
    return products, nil
}