package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type OrganizationRepo struct{}

func NewOrganizationRepo() *OrganizationRepo { return &OrganizationRepo{} }

func (r *OrganizationRepo) CreateOrg(ctx context.Context, o *model.Organization) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO organizations (name, code, industry, plan, max_users, max_rooms)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_at`,
		o.Name, o.Code, o.Industry, o.Plan, o.MaxUsers, o.MaxRooms,
	).Scan(&o.ID, &o.CreatedAt)
}

func (r *OrganizationRepo) GetOrg(ctx context.Context, id int64) (*model.Organization, error) {
	var o model.Organization
	err := database.Get().QueryRow(ctx,
		`SELECT id, name, code, industry, plan, max_users, max_rooms, status, expired_at, created_at, updated_at
		FROM organizations WHERE id=$1`, id).Scan(
		&o.ID, &o.Name, &o.Code, &o.Industry, &o.Plan, &o.MaxUsers, &o.MaxRooms, &o.Status, &o.ExpiredAt, &o.CreatedAt, &o.UpdatedAt)
	return &o, err
}

func (r *OrganizationRepo) ListOrgs(ctx context.Context, page, pageSize int) ([]model.Organization, int, error) {
	var total int
	database.Get().QueryRow(ctx, "SELECT COUNT(*) FROM organizations").Scan(&total)
	offset := (page - 1) * pageSize
	rows, err := database.Get().Query(ctx,
		`SELECT id, name, code, industry, plan, max_users, max_rooms, status, created_at
		FROM organizations ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.Organization
	for rows.Next() {
		var o model.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Code, &o.Industry, &o.Plan, &o.MaxUsers, &o.MaxRooms, &o.Status, &o.CreatedAt); err != nil {
			continue
		}
		list = append(list, o)
	}
	return list, total, nil
}

func (r *OrganizationRepo) UpdateOrg(ctx context.Context, o *model.Organization) error {
	_, err := database.Get().Exec(ctx,
		`UPDATE organizations SET name=$1, industry=$2, plan=$3, max_users=$4, max_rooms=$5 WHERE id=$6`,
		o.Name, o.Industry, o.Plan, o.MaxUsers, o.MaxRooms, o.ID)
	return err
}

func (r *OrganizationRepo) CreateDept(ctx context.Context, d *model.Department) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO departments (organization_id, parent_id, name, code, leader_id, sort_order, path, level)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at`,
		d.OrganizationID, d.ParentID, d.Name, d.Code, d.LeaderID, d.SortOrder, d.Path, d.Level,
	).Scan(&d.ID, &d.CreatedAt)
}

func (r *OrganizationRepo) GetDeptTree(ctx context.Context, orgID int64) ([]model.Department, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, organization_id, parent_id, name, code, leader_id, sort_order, path, level, status, created_at
		FROM departments WHERE organization_id=$1 AND status='active'
		ORDER BY level, sort_order`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var all []model.Department
	for rows.Next() {
		var d model.Department
		if err := rows.Scan(&d.ID, &d.OrganizationID, &d.ParentID, &d.Name, &d.Code, &d.LeaderID, &d.SortOrder, &d.Path, &d.Level, &d.Status, &d.CreatedAt); err != nil {
			continue
		}
		all = append(all, d)
	}
	// Build tree
	return buildDeptTree(all, nil), nil
}

func (r *OrganizationRepo) UpdateDept(ctx context.Context, d *model.Department) error {
	_, err := database.Get().Exec(ctx,
		`UPDATE departments SET name=$1, leader_id=$2, sort_order=$3 WHERE id=$4`,
		d.Name, d.LeaderID, d.SortOrder, d.ID)
	return err
}

func (r *OrganizationRepo) DeleteDept(ctx context.Context, id int64) error {
	_, err := database.Get().Exec(ctx, "UPDATE departments SET status='deleted' WHERE id=$1", id)
	return err
}

func buildDeptTree(all []model.Department, parentID *int64) []model.Department {
	var tree []model.Department
	for _, d := range all {
		if (parentID == nil && d.ParentID == nil) || (parentID != nil && d.ParentID != nil && *d.ParentID == *parentID) {
			pid := d.ID
			d.Children = buildDeptTree(all, &pid)
			tree = append(tree, d)
		}
	}
	return tree
}

// ─── RBAC Repository ───

type RBACRepo struct{}

func NewRBACRepo() *RBACRepo { return &RBACRepo{} }

func (r *RBACRepo) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, code, name, module, resource, action FROM permissions ORDER BY module, resource, action`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Permission
	for rows.Next() {
		var p model.Permission
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Module, &p.Resource, &p.Action); err != nil {
			continue
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *RBACRepo) CreateRole(ctx context.Context, role *model.Role) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO roles (organization_id, name, code, description)
		VALUES ($1,$2,$3,$4) RETURNING id, created_at`,
		role.OrganizationID, role.Name, role.Code, role.Description,
	).Scan(&role.ID, &role.CreatedAt)
}

func (r *RBACRepo) ListRoles(ctx context.Context, orgID int64) ([]model.Role, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, organization_id, name, code, is_system, description, created_at
		FROM roles WHERE organization_id=$1 ORDER BY created_at`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Role
	for rows.Next() {
		var role model.Role
		if err := rows.Scan(&role.ID, &role.OrganizationID, &role.Name, &role.Code, &role.IsSystem, &role.Description, &role.CreatedAt); err != nil {
			continue
		}
		list = append(list, role)
	}
	return list, nil
}

func (r *RBACRepo) UpdateRolePermissions(ctx context.Context, roleID int64, permIDs []int64) error {
	tx, err := database.Get().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tx.Exec(ctx, "DELETE FROM role_permissions WHERE role_id=$1", roleID)
	for _, pid := range permIDs {
		if _, err := tx.Exec(ctx, "INSERT INTO role_permissions (role_id, permission_id) VALUES ($1,$2)", roleID, pid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *RBACRepo) AssignUserRole(ctx context.Context, userID, roleID int64) error {
	_, err := database.Get().Exec(ctx,
		"INSERT INTO user_roles (user_id, role_id) VALUES ($1,$2) ON CONFLICT DO NOTHING", userID, roleID)
	return err
}

func (r *RBACRepo) GetUserPermissions(ctx context.Context, userID int64) ([]model.Permission, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT DISTINCT p.id, p.code, p.name, p.module, p.resource, p.action
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Permission
	for rows.Next() {
		var p model.Permission
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Module, &p.Resource, &p.Action); err != nil {
			continue
		}
		list = append(list, p)
	}
	return list, nil
}

// ─── Audit Repository ───

type AuditRepo struct{}

func NewAuditRepo() *AuditRepo { return &AuditRepo{} }

func (r *AuditRepo) Create(ctx context.Context, log *model.AuditLog) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO audit_logs (user_id, username, action, resource, resource_id, detail, ip, user_agent, request_id, duration, status_code)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id, created_at`,
		log.UserID, log.Username, log.Action, log.Resource, log.ResourceID, log.Detail,
		log.IP, log.UserAgent, log.RequestID, log.Duration, log.StatusCode,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *AuditRepo) List(ctx context.Context, page, pageSize int, filters map[string]string) ([]model.AuditLog, int, error) {
	var conds []string
	var args []interface{}
	idx := 1
	addF := func(col, val string) {
		conds = append(conds, fmt.Sprintf("%s=$%d", col, idx))
		args = append(args, val)
		idx++
	}
	if v, ok := filters["user_id"]; ok && v != "" {
		addF("user_id", v)
	}
	if v, ok := filters["action"]; ok && v != "" {
		addF("action", v)
	}
	if v, ok := filters["resource"]; ok && v != "" {
		addF("resource", v)
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	var total int
	database.Get().QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs "+where, args...).Scan(&total)
	offset := (page - 1) * pageSize
	query := `SELECT id, user_id, username, action, resource, resource_id, detail, ip, user_agent, request_id, duration, status_code, created_at
		FROM audit_logs ` + where + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, pageSize, offset)
	rows, err := database.Get().Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Username, &l.Action, &l.Resource, &l.ResourceID, &l.Detail, &l.IP, &l.UserAgent, &l.RequestID, &l.Duration, &l.StatusCode, &l.CreatedAt); err != nil {
			continue
		}
		list = append(list, l)
	}
	return list, total, nil
}

func (r *AuditRepo) GetStats(ctx context.Context, startDate, endDate string) (map[string]interface{}, error) {
	args := []interface{}{}
	where := ""
	if startDate != "" && endDate != "" {
		where = "WHERE created_at BETWEEN $1 AND $2"
		args = append(args, startDate, endDate+" 23:59:59")
	}
	rows, err := database.Get().Query(ctx,
		`SELECT action, COUNT(*) FROM audit_logs `+where+` GROUP BY action ORDER BY COUNT(*) DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var actions []map[string]interface{}
	var total int64
	for rows.Next() {
		var action string
		var count int64
		if err := rows.Scan(&action, &count); err != nil {
			continue
		}
		actions = append(actions, map[string]interface{}{"action": action, "count": count})
		total += count
	}
	return map[string]interface{}{"total": total, "by_action": actions}, nil
}

// ─── Export Task Repository ───

type ExportRepo struct{}

func NewExportRepo() *ExportRepo { return &ExportRepo{} }

func (r *ExportRepo) Create(ctx context.Context, t *model.ExportTask) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO export_tasks (user_id, type, format, params, status)
		VALUES ($1,$2,$3,$4,'pending') RETURNING id, created_at`,
		t.UserID, t.Type, t.Format, t.Params,
	).Scan(&t.ID, &t.CreatedAt)
}

func (r *ExportRepo) GetByID(ctx context.Context, id int64) (*model.ExportTask, error) {
	var t model.ExportTask
	err := database.Get().QueryRow(ctx,
		`SELECT id, user_id, type, format, params, status, progress, total_rows, file_path, file_size, error_msg, started_at, completed_at, expires_at, created_at
		FROM export_tasks WHERE id=$1`, id).Scan(
		&t.ID, &t.UserID, &t.Type, &t.Format, &t.Params, &t.Status, &t.Progress, &t.TotalRows,
		&t.FilePath, &t.FileSize, &t.ErrorMsg, &t.StartedAt, &t.CompletedAt, &t.ExpiresAt, &t.CreatedAt)
	return &t, err
}

func (r *ExportRepo) ListByUser(ctx context.Context, userID int64, limit int) ([]model.ExportTask, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := database.Get().Query(ctx,
		`SELECT id, user_id, type, format, status, progress, total_rows, file_path, created_at
		FROM export_tasks WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.ExportTask
	for rows.Next() {
		var t model.ExportTask
		if err := rows.Scan(&t.ID, &t.UserID, &t.Type, &t.Format, &t.Status, &t.Progress, &t.TotalRows, &t.FilePath, &t.CreatedAt); err != nil {
			continue
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *ExportRepo) UpdateStatus(ctx context.Context, id int64, status, filePath, errMsg string) error {
	_, err := database.Get().Exec(ctx,
		`UPDATE export_tasks SET status=$1, file_path=COALESCE($2,file_path), error_msg=COALESCE($3,error_msg) WHERE id=$4`,
		status, filePath, errMsg, id)
	return err
}

// ─── Event Tracking Repository ───

type EventRepo struct{}

func NewEventRepo() *EventRepo { return &EventRepo{} }

func (r *EventRepo) Track(ctx context.Context, e *model.TrackEvent) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO track_events (event_name, platform, user_id, session_id, live_room_id, product_id, properties, timestamp)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at`,
		e.EventName, e.Platform, e.UserID, e.SessionID, e.LiveRoomID, e.ProductID, e.Properties, e.Timestamp,
	).Scan(&e.ID, &e.CreatedAt)
}

func (r *EventRepo) BatchTrack(ctx context.Context, events []model.TrackEvent) error {
	tx, err := database.Get().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, e := range events {
		_, err := tx.Exec(ctx,
			`INSERT INTO track_events (event_name, platform, user_id, session_id, live_room_id, product_id, properties, timestamp)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			e.EventName, e.Platform, e.UserID, e.SessionID, e.LiveRoomID, e.ProductID, e.Properties, e.Timestamp)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *EventRepo) GetFunnelEvents(ctx context.Context, liveRoomID int64) ([]model.TrackEvent, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, event_name, platform, user_id, session_id, live_room_id, product_id, properties, timestamp
		FROM track_events WHERE live_room_id=$1
		ORDER BY timestamp ASC LIMIT 1000`, liveRoomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.TrackEvent
	for rows.Next() {
		var e model.TrackEvent
		if err := rows.Scan(&e.ID, &e.EventName, &e.Platform, &e.UserID, &e.SessionID, &e.LiveRoomID, &e.ProductID, &e.Properties, &e.Timestamp); err != nil {
			continue
		}
		list = append(list, e)
	}
	return list, nil
}

// ─── Data Quality Repository ───

type DataQualityRepo struct{}

func NewDataQualityRepo() *DataQualityRepo { return &DataQualityRepo{} }

func (r *DataQualityRepo) CreateRule(ctx context.Context, rule *model.DataQualityRule) error {
	return database.Get().QueryRow(ctx,
		`INSERT INTO data_quality_rules (name, table_name, column_name, rule_type, expression, severity, is_enabled, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, created_at`,
		rule.Name, rule.TableName, rule.ColumnName, rule.RuleType, rule.Expression, rule.Severity, rule.IsEnabled, rule.CreatedBy,
	).Scan(&rule.ID, &rule.CreatedAt)
}

func (r *DataQualityRepo) ListRules(ctx context.Context) ([]model.DataQualityRule, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, name, table_name, column_name, rule_type, expression, severity, is_enabled, last_check_at, created_by, created_at
		FROM data_quality_rules ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.DataQualityRule
	for rows.Next() {
		var rule model.DataQualityRule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.TableName, &rule.ColumnName, &rule.RuleType, &rule.Expression, &rule.Severity, &rule.IsEnabled, &rule.LastCheckAt, &rule.CreatedBy, &rule.CreatedAt); err != nil {
			continue
		}
		list = append(list, rule)
	}
	return list, nil
}

// ─── Metrics Aggregation Repository ───

type MetricsAggRepo struct{}

func NewMetricsAggRepo() *MetricsAggRepo { return &MetricsAggRepo{} }

func (r *MetricsAggRepo) GetDailyMetrics(ctx context.Context, platform, startDate, endDate string) ([]model.MetricsDaily, error) {
	args := []interface{}{startDate, endDate + " 23:59:59"}
	platformFilter := ""
	if platform != "" {
		platformFilter = " AND platform=$3"
		args = append(args, platform)
	}
	rows, err := database.Get().Query(ctx,
		`SELECT id, platform, date, live_room_count, gmv, order_count, viewer_count, peak_viewers,
		like_count, comment_count, share_count, conversion_rate, avg_watch_time, new_followers,
		gift_count, gift_value, refund_amount, commission, created_at
		FROM metrics_daily WHERE date BETWEEN $1 AND $2`+platformFilter+" ORDER BY date DESC", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.MetricsDaily
	for rows.Next() {
		var m model.MetricsDaily
		if err := rows.Scan(&m.ID, &m.Platform, &m.Date, &m.LiveRoomCount, &m.GMV, &m.OrderCount, &m.ViewerCount, &m.PeakViewers,
			&m.LikeCount, &m.CommentCount, &m.ShareCount, &m.ConversionRate, &m.AvgWatchTime, &m.NewFollowers,
			&m.GiftCount, &m.GiftValue, &m.RefundAmount, &m.Commission, &m.CreatedAt); err != nil {
			continue
		}
		list = append(list, m)
	}
	return list, nil
}
