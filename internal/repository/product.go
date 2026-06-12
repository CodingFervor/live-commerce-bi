package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type ProductRepo struct{}

func NewProductRepo() *ProductRepo { return &ProductRepo{} }

func (r *ProductRepo) Create(ctx context.Context, p *model.Product) error {
	query := `INSERT INTO products (name, platform, platform_product_id, category, brand,
		price, original_price, live_price, image_url, description, tags)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id, created_at`
	return database.Get().QueryRow(ctx, query,
		p.Name, p.Platform, p.PlatformProductID, p.Category, p.Brand,
		p.Price, p.OriginalPrice, p.LivePrice, p.ImageURL, p.Description, p.Tags,
	).Scan(&p.ID, &p.CreatedAt)
}

func (r *ProductRepo) List(ctx context.Context, page, pageSize int, platform, category string) ([]model.Product, int, error) {
	var conds []string
	var args []interface{}
	idx := 1
	if platform != "" {
		conds = append(conds, fmt.Sprintf("platform=$%d", idx))
		args = append(args, platform)
		idx++
	}
	if category != "" {
		conds = append(conds, fmt.Sprintf("category=$%d", idx))
		args = append(args, category)
		idx++
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	var total int
	database.Get().QueryRow(ctx, "SELECT COUNT(*) FROM products "+where, args...).Scan(&total)
	offset := (page - 1) * pageSize
	q := `SELECT id, name, platform, platform_product_id, category, brand, price, original_price,
		live_price, image_url, description, tags, status, created_at, updated_at
		FROM products ` + where + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, pageSize, offset)
	rows, err := database.Get().Query(ctx, q, args...)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var list []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Platform, &p.PlatformProductID, &p.Category, &p.Brand,
			&p.Price, &p.OriginalPrice, &p.LivePrice, &p.ImageURL, &p.Description, &p.Tags, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, total, nil
}

func (r *ProductRepo) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	var p model.Product
	err := database.Get().QueryRow(ctx,
		`SELECT id, name, platform, platform_product_id, category, brand, price, original_price,
		live_price, image_url, description, tags, status, created_at, updated_at
		FROM products WHERE id=$1`, id).Scan(
		&p.ID, &p.Name, &p.Platform, &p.PlatformProductID, &p.Category, &p.Brand,
		&p.Price, &p.OriginalPrice, &p.LivePrice, &p.ImageURL, &p.Description, &p.Tags, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

func (r *ProductRepo) Update(ctx context.Context, p *model.Product) error {
	_, err := database.Get().Exec(ctx,
		`UPDATE products SET name=$1, category=$2, brand=$3, price=$4, live_price=$5, status=$6 WHERE id=$7`,
		p.Name, p.Category, p.Brand, p.Price, p.LivePrice, p.Status, p.ID)
	return err
}

func (r *ProductRepo) GetRankings(ctx context.Context, metric string, limit int, category string) ([]model.ProductRanking, error) {
	orderCol := "total_revenue"
	switch metric {
	case "sold":
		orderCol = "total_sold"
	case "conversion":
		orderCol = "avg_conversion"
	}
	args := []interface{}{limit}
	idx := 1
	where := ""
	if category != "" {
		where = fmt.Sprintf("WHERE p.category = $%d", idx+1)
		args = append(args, category)
	}
	query := fmt.Sprintf(`SELECT p.id, p.name, p.brand, p.category, p.image_url, p.price,
		COALESCE(SUM(lrp.orders),0) AS total_sold,
		COALESCE(SUM(lrp.revenue),0) AS total_revenue,
		COUNT(DISTINCT lrp.live_room_id) AS appearances,
		CASE WHEN SUM(lrp.clicks) > 0 THEN SUM(lrp.orders)::FLOAT / SUM(lrp.clicks) ELSE 0 END AS avg_conversion
		FROM products p LEFT JOIN live_room_products lrp ON p.id = lrp.product_id
		%s GROUP BY p.id ORDER BY %s DESC LIMIT $1`, where, orderCol)
	rows, err := database.Get().Query(ctx, query, args...)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.ProductRanking
	rank := 1
	for rows.Next() {
		var pr model.ProductRanking
		if err := rows.Scan(&pr.ProductID, &pr.ProductName, &pr.Brand, &pr.Category, &pr.ImageURL, &pr.Price,
			&pr.TotalSold, &pr.TotalRevenue, &pr.Appearances, &pr.ConversionRate); err != nil {
			return nil, err
		}
		pr.Rank = rank
		list = append(list, pr)
		rank++
	}
	return list, nil
}
