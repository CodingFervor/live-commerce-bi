package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type OrderRepo struct{}

func NewOrderRepo() *OrderRepo { return &OrderRepo{} }

func (r *OrderRepo) List(ctx context.Context, page, pageSize int, filters map[string]string) ([]model.Order, int, error) {
	var conds []string
	var args []interface{}
	idx := 1
	addFilter := func(col, val string) {
		conds = append(conds, fmt.Sprintf("o.%s=$%d", col, idx))
		args = append(args, val)
		idx++
	}
	if v, ok := filters["platform"]; ok && v != "" { addFilter("platform", v) }
	if v, ok := filters["status"]; ok && v != "" { addFilter("status", v) }
	if v, ok := filters["live_room_id"]; ok && v != "" { addFilter("live_room_id", v) }
	if v, ok := filters["streamer_id"]; ok && v != "" { addFilter("streamer_id", v) }
	if v, ok := filters["start_date"]; ok && v != "" {
		conds = append(conds, fmt.Sprintf("o.created_at >= $%d", idx))
		args = append(args, v)
		idx++
	}
	if v, ok := filters["end_date"]; ok && v != "" {
		conds = append(conds, fmt.Sprintf("o.created_at <= $%d", idx))
		args = append(args, v+" 23:59:59")
		idx++
	}
	where := ""
	if len(conds) > 0 { where = "WHERE " + strings.Join(conds, " AND ") }

	var total int
	database.Get().QueryRow(ctx, "SELECT COUNT(*) FROM orders o "+where, args...).Scan(&total)
	offset := (page - 1) * pageSize
	q := `SELECT o.id, o.order_no, o.platform, o.platform_order_id, o.live_room_id, o.product_id,
		o.streamer_id, o.buyer_id, o.quantity, o.unit_price, o.total_amount,
		o.discount_amount, o.actual_amount, o.commission_rate, o.commission,
		o.status, o.paid_at, o.shipped_at, o.completed_at, o.refunded_at,
		o.created_at, o.updated_at,
		p.name AS product_name, s.name AS streamer_name, lr.title AS room_title
		FROM orders o
		LEFT JOIN products p ON o.product_id = p.id
		LEFT JOIN streamers s ON o.streamer_id = s.id
		LEFT JOIN live_rooms lr ON o.live_room_id = lr.id
		` + where + fmt.Sprintf(" ORDER BY o.created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, pageSize, offset)
	rows, err := database.Get().Query(ctx, q, args...)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var list []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.Platform, &o.PlatformOrderID, &o.LiveRoomID, &o.ProductID,
			&o.StreamerID, &o.BuyerID, &o.Quantity, &o.UnitPrice, &o.TotalAmount,
			&o.DiscountAmount, &o.ActualAmount, &o.CommissionRate, &o.Commission,
			&o.Status, &o.PaidAt, &o.ShippedAt, &o.CompletedAt, &o.RefundedAt,
			&o.CreatedAt, &o.UpdatedAt, &o.ProductName, &o.StreamerName, &o.RoomTitle); err != nil {
			return nil, 0, err
		}
		list = append(list, o)
	}
	return list, total, nil
}

func (r *OrderRepo) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	var o model.Order
	err := database.Get().QueryRow(ctx,
		`SELECT id, order_no, platform, platform_order_id, live_room_id, product_id,
		streamer_id, buyer_id, quantity, unit_price, total_amount,
		discount_amount, actual_amount, commission_rate, commission,
		status, paid_at, shipped_at, completed_at, refunded_at, created_at, updated_at
		FROM orders WHERE id=$1`, id).Scan(
		&o.ID, &o.OrderNo, &o.Platform, &o.PlatformOrderID, &o.LiveRoomID, &o.ProductID,
		&o.StreamerID, &o.BuyerID, &o.Quantity, &o.UnitPrice, &o.TotalAmount,
		&o.DiscountAmount, &o.ActualAmount, &o.CommissionRate, &o.Commission,
		&o.Status, &o.PaidAt, &o.ShippedAt, &o.CompletedAt, &o.RefundedAt, &o.CreatedAt, &o.UpdatedAt)
	return &o, err
}

func (r *OrderRepo) GetStats(ctx context.Context, startDate, endDate string) (*model.OrderStats, error) {
	args := []interface{}{}
	where := ""
	if startDate != "" && endDate != "" {
		where = "WHERE created_at BETWEEN $1 AND $2"
		args = append(args, startDate, endDate+" 23:59:59")
	}
	var s model.OrderStats
	q := `SELECT COUNT(*) AS total_orders,
		COALESCE(SUM(total_amount),0),
		COALESCE(SUM(actual_amount),0),
		COALESCE(SUM(commission),0),
		COALESCE(SUM(CASE WHEN status='refunded' THEN actual_amount ELSE 0 END),0),
		CASE WHEN COUNT(*) > 0 THEN COALESCE(SUM(actual_amount),0)/COUNT(*)::DECIMAL ELSE 0 END,
		CASE WHEN COUNT(*) > 0 THEN COUNT(CASE WHEN status='completed' THEN 1 END)::DECIMAL/COUNT(*)::DECIMAL ELSE 0 END,
		CASE WHEN COUNT(*) > 0 THEN COUNT(CASE WHEN status='refunded' THEN 1 END)::DECIMAL/COUNT(*)::DECIMAL ELSE 0 END
		FROM orders ` + where
	err := database.Get().QueryRow(ctx, q, args...).Scan(
		&s.TotalOrders, &s.TotalGMV, &s.TotalActual, &s.TotalCommission, &s.TotalRefund,
		&s.AvgOrderValue, &s.CompletedRate, &s.RefundRate)
	return &s, err
}

func (r *OrderRepo) GetRevenueByDate(ctx context.Context, startDate, endDate, groupBy string) ([]model.RevenueRecord, error) {
	q := `SELECT date, platform, gmv, actual_revenue, commission, refund_amount, order_count, avg_order_value
		FROM revenue_records WHERE date BETWEEN $1 AND $2 ORDER BY date DESC`
	args := []interface{}{startDate, endDate}
	if groupBy == "platform" {
		q = `SELECT date, platform, SUM(gmv), SUM(actual_revenue), SUM(commission), SUM(refund_amount), SUM(order_count), AVG(avg_order_value)
			FROM revenue_records WHERE date BETWEEN $1 AND $2 GROUP BY date, platform ORDER BY date DESC, platform`
	}
	rows, err := database.Get().Query(ctx, q, args...)
	if err != nil { return nil, err }
	defer rows.Close()
	var list []model.RevenueRecord
	for rows.Next() {
		var rec model.RevenueRecord
		if err := rows.Scan(&rec.Date, &rec.Platform, &rec.GMV, &rec.ActualRevenue, &rec.Commission, &rec.RefundAmount, &rec.OrderCount, &rec.AvgOrderValue); err != nil {
			return nil, err
		}
		list = append(list, rec)
	}
	return list, nil
}
