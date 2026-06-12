package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/pkg/logger"
)

// ═══ Enhanced Service Layer ═══
// Async Export Executor, Data Quality Checker, Schedule Dispatcher

// ─── Async Export Executor ───

type ExportExecutor struct {
	mu     sync.Mutex
	tasks  map[int64]context.CancelFunc
	engine *ExportEngine
}

func NewExportExecutor() *ExportExecutor {
	return &ExportExecutor{
		tasks:  make(map[int64]context.CancelFunc),
		engine: NewExportEngine(),
	}
}

// StartExport begins asynchronous export execution
func (e *ExportExecutor) StartExport(ctx context.Context, task *model.ExportTask) error {
	exportCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)

	e.mu.Lock()
	e.tasks[task.ID] = cancel
	e.mu.Unlock()

	go e.executeExport(exportCtx, task)
	return nil
}

func (e *ExportExecutor) executeExport(ctx context.Context, task *model.ExportTask) {
	defer func() {
		e.mu.Lock()
		delete(e.tasks, task.ID)
		e.mu.Unlock()
	}()

	// Mark as running
	now := time.Now()
	task.StartedAt = &now
	task.Status = "running"
	updateExportStatus(task.ID, "running", "", "")

	// Parse params
	var params struct {
		SQL      string   `json:"sql"`
		Headers  []string `json:"headers"`
		Filename string   `json:"filename"`
	}
	if err := json.Unmarshal([]byte(task.Params), &params); err != nil {
		task.Status = "failed"
		task.ErrorMsg = "invalid params: " + err.Error()
		updateExportStatus(task.ID, "failed", "", task.ErrorMsg)
		return
	}

	// Query data
	rows, err := database.Get().Query(ctx, params.SQL)
	if err != nil {
		task.Status = "failed"
		task.ErrorMsg = "query failed: " + err.Error()
		updateExportStatus(task.ID, "failed", "", task.ErrorMsg)
		return
	}
	defer rows.Close()

	cols := rows.FieldDescriptions()
	if len(params.Headers) == 0 {
		params.Headers = make([]string, len(cols))
		for i, col := range cols {
			params.Headers[i] = string(col.Name)
		}
	}

	var allRows [][]string
	rowCount := 0
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			continue
		}
		row := make([]string, len(vals))
		for i, v := range vals {
			row[i] = fmt.Sprintf("%v", v)
		}
		allRows = append(allRows, row)
		rowCount++

		// Update progress every 1000 rows
		if rowCount%1000 == 0 {
			progress := min(90, rowCount/100)
			updateExportProgress(task.ID, progress, rowCount)
		}
	}

	// Build output
	var content string
	switch task.Format {
	case "csv":
		content = e.engine.BuildCSV(params.Headers, allRows)
	case "json":
		content = e.engine.BuildJSON(params.Headers, allRows)
	default:
		content = e.engine.BuildCSV(params.Headers, allRows)
	}

	// Save to file
	exportDir := "exports"
	os.MkdirAll(exportDir, 0755)
	filename := fmt.Sprintf("export_%d_%s.%s", task.ID, time.Now().Format("20060102_150405"), task.Format)
	filePath := filepath.Join(exportDir, filename)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		task.Status = "failed"
		task.ErrorMsg = "write file failed: " + err.Error()
		updateExportStatus(task.ID, "failed", "", task.ErrorMsg)
		return
	}

	fileInfo, _ := os.Stat(filePath)
	fileSize := int64(0)
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	// Mark as completed
	completedAt := time.Now()
	expiresAt := completedAt.Add(24 * time.Hour)
	task.Status = "completed"
	task.FilePath = filePath
	task.FileSize = fileSize
	task.TotalRows = rowCount
	task.CompletedAt = &completedAt
	task.ExpiresAt = &expiresAt

	updateExportStatus(task.ID, "completed", filePath, "")
	updateExportProgress(task.ID, 100, rowCount)

	logger.Info("Export task %d completed: %d rows, %s", task.ID, rowCount, filePath)
}

// CancelExport cancels a running export task
func (e *ExportExecutor) CancelExport(taskID int64) bool {
	e.mu.Lock()
	cancel, ok := e.tasks[taskID]
	e.mu.Unlock()

	if ok {
		cancel()
		updateExportStatus(taskID, "cancelled", "", "cancelled by user")
		return true
	}
	return false
}

func updateExportStatus(id int64, status, filePath, errMsg string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	database.Get().Exec(ctx,
		`UPDATE export_tasks SET status=$1, file_path=COALESCE($2,file_path), error_msg=COALESCE($3,error_msg) WHERE id=$4`,
		status, filePath, errMsg, id)
}

func updateExportProgress(id int64, progress, totalRows int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	database.Get().Exec(ctx,
		`UPDATE export_tasks SET progress=$1, total_rows=$2 WHERE id=$3`,
		progress, totalRows, id)
}

// ─── Data Quality Checker ───

type DataQualityChecker struct{}

func NewDataQualityChecker() *DataQualityChecker { return &DataQualityChecker{} }

type QualityCheckResult struct {
	RuleID    int64   `json:"rule_id"`
	RuleName  string  `json:"rule_name"`
	TableName string  `json:"table_name"`
	Status    string  `json:"status"`
	TotalRows int     `json:"total_rows"`
	PassRows  int     `json:"pass_rows"`
	FailRows  int     `json:"fail_rows"`
	PassRate  float64 `json:"pass_rate"`
	Detail    string  `json:"detail"`
}

// RunCheck executes a data quality rule against the database
func (c *DataQualityChecker) RunCheck(ctx context.Context, rule *model.DataQualityRule) (*QualityCheckResult, error) {
	result := &QualityCheckResult{
		RuleID:    rule.ID,
		RuleName:  rule.Name,
		TableName: rule.TableName,
	}

	switch rule.RuleType {
	case "not_null":
		result = c.checkNotNull(ctx, rule)
	case "unique":
		result = c.checkUnique(ctx, rule)
	case "range":
		result = c.checkRange(ctx, rule)
	case "regex":
		result = c.checkRegex(ctx, rule)
	case "custom":
		result = c.checkCustom(ctx, rule)
	default:
		result.Status = "error"
		result.Detail = fmt.Sprintf("unsupported rule type: %s", rule.RuleType)
	}

	// Save result to database
	c.saveResult(ctx, result)

	// Update rule last_check_at
	database.Get().Exec(ctx,
		"UPDATE data_quality_rules SET last_check_at=NOW() WHERE id=$1", rule.ID)

	return result, nil
}

func (c *DataQualityChecker) checkNotNull(ctx context.Context, rule *model.DataQualityRule) *QualityCheckResult {
	result := &QualityCheckResult{RuleID: rule.ID, RuleName: rule.Name, TableName: rule.TableName}

	// Count total rows
	database.Get().QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s", safeTableName(rule.TableName)),
	).Scan(&result.TotalRows)

	// Count null rows
	database.Get().QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s IS NULL", safeTableName(rule.TableName), safeColumnName(rule.ColumnName)),
	).Scan(&result.FailRows)

	result.PassRows = result.TotalRows - result.FailRows
	if result.TotalRows > 0 {
		result.PassRate = float64(result.PassRows) / float64(result.TotalRows) * 100
	}
	if result.PassRate >= 99.9 {
		result.Status = "pass"
	} else if result.PassRate >= 95 {
		result.Status = "warning"
	} else {
		result.Status = "fail"
	}
	result.Detail = fmt.Sprintf("NULL values in %s.%s: %d/%d", rule.TableName, rule.ColumnName, result.FailRows, result.TotalRows)

	return result
}

func (c *DataQualityChecker) checkUnique(ctx context.Context, rule *model.DataQualityRule) *QualityCheckResult {
	result := &QualityCheckResult{RuleID: rule.ID, RuleName: rule.Name, TableName: rule.TableName}

	database.Get().QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s", safeTableName(rule.TableName)),
	).Scan(&result.TotalRows)

	var duplicateCount int
	database.Get().QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM (SELECT %s, COUNT(*) as c FROM %s GROUP BY %s HAVING COUNT(*) > 1) t",
			safeColumnName(rule.ColumnName), safeTableName(rule.TableName), safeColumnName(rule.ColumnName)),
	).Scan(&duplicateCount)

	result.FailRows = duplicateCount
	result.PassRows = result.TotalRows - duplicateCount
	if result.TotalRows > 0 {
		result.PassRate = float64(result.PassRows) / float64(result.TotalRows) * 100
	}
	if result.PassRate >= 99.9 {
		result.Status = "pass"
	} else if result.PassRate >= 95 {
		result.Status = "warning"
	} else {
		result.Status = "fail"
	}
	result.Detail = fmt.Sprintf("Duplicate values in %s.%s: %d", rule.TableName, rule.ColumnName, duplicateCount)

	return result
}

func (c *DataQualityChecker) checkRange(ctx context.Context, rule *model.DataQualityRule) *QualityCheckResult {
	result := &QualityCheckResult{RuleID: rule.ID, RuleName: rule.Name, TableName: rule.TableName}

	database.Get().QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s", safeTableName(rule.TableName)),
	).Scan(&result.TotalRows)

	// Expression is like "min,max"
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE NOT (%s >= %s)",
		safeTableName(rule.TableName), safeColumnName(rule.ColumnName), rule.Expression)

	database.Get().QueryRow(ctx, query).Scan(&result.FailRows)

	result.PassRows = result.TotalRows - result.FailRows
	if result.TotalRows > 0 {
		result.PassRate = float64(result.PassRows) / float64(result.TotalRows) * 100
	}
	if result.PassRate >= 99 {
		result.Status = "pass"
	} else if result.PassRate >= 90 {
		result.Status = "warning"
	} else {
		result.Status = "fail"
	}
	result.Detail = fmt.Sprintf("Out-of-range values in %s.%s: %d/%d", rule.TableName, rule.ColumnName, result.FailRows, result.TotalRows)

	return result
}

func (c *DataQualityChecker) checkRegex(ctx context.Context, rule *model.DataQualityRule) *QualityCheckResult {
	result := &QualityCheckResult{RuleID: rule.ID, RuleName: rule.Name, TableName: rule.TableName}

	database.Get().QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s", safeTableName(rule.TableName)),
	).Scan(&result.TotalRows)

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s !~ '%s'",
		safeTableName(rule.TableName), safeColumnName(rule.ColumnName), rule.Expression)

	database.Get().QueryRow(ctx, query).Scan(&result.FailRows)

	result.PassRows = result.TotalRows - result.FailRows
	if result.TotalRows > 0 {
		result.PassRate = float64(result.PassRows) / float64(result.TotalRows) * 100
	}
	if result.PassRate >= 99 {
		result.Status = "pass"
	} else if result.PassRate >= 90 {
		result.Status = "warning"
	} else {
		result.Status = "fail"
	}
	result.Detail = fmt.Sprintf("Regex mismatch in %s.%s: %d/%d", rule.TableName, rule.ColumnName, result.FailRows, result.TotalRows)

	return result
}

func (c *DataQualityChecker) checkCustom(ctx context.Context, rule *model.DataQualityRule) *QualityCheckResult {
	result := &QualityCheckResult{RuleID: rule.ID, RuleName: rule.Name, TableName: rule.TableName}

	database.Get().QueryRow(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s", safeTableName(rule.TableName)),
	).Scan(&result.TotalRows)

	// Expression is a WHERE clause for failing rows
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s",
		safeTableName(rule.TableName), rule.Expression)

	database.Get().QueryRow(ctx, query).Scan(&result.FailRows)

	result.PassRows = result.TotalRows - result.FailRows
	if result.TotalRows > 0 {
		result.PassRate = float64(result.PassRows) / float64(result.TotalRows) * 100
	}
	if result.PassRate >= 99 {
		result.Status = "pass"
	} else if result.PassRate >= 90 {
		result.Status = "warning"
	} else {
		result.Status = "fail"
	}
	result.Detail = fmt.Sprintf("Custom check on %s: %d/%d passed", rule.TableName, result.PassRows, result.TotalRows)

	return result
}

func (c *DataQualityChecker) saveResult(ctx context.Context, r *QualityCheckResult) {
	detail := r.Detail
	if len(detail) > 500 {
		detail = detail[:500]
	}
	database.Get().Exec(ctx,
		`INSERT INTO data_quality_results (rule_id, status, total_rows, pass_rows, fail_rows, pass_rate, detail, checked_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())`,
		r.RuleID, r.Status, r.TotalRows, r.PassRows, r.FailRows, r.PassRate, detail)
}

// safeTableName validates table name to prevent SQL injection
func safeTableName(name string) string {
	allowed := map[string]bool{
		"orders": true, "live_rooms": true, "products": true, "streamers": true,
		"users": true, "viewer_metrics": true, "viewer_demographics": true,
		"revenue_records": true, "platform_metrics": true, "data_sources": true,
		"track_events": true, "export_tasks": true, "audit_logs": true,
		"metrics_hourly": true, "metrics_daily": true,
	}
	if allowed[name] {
		return name
	}
	return "INVALID_TABLE"
}

// safeColumnName validates column name
func safeColumnName(name string) string {
	if name == "" {
		return "id"
	}
	// Only allow alphanumeric and underscore
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return "id"
		}
	}
	return name
}

// ─── Schedule Dispatcher ───

type ScheduleDispatcher struct {
	tasks   map[int64]context.CancelFunc
	mu      sync.Mutex
	running bool
}

func NewScheduleDispatcher() *ScheduleDispatcher {
	return &ScheduleDispatcher{
		tasks: make(map[int64]context.CancelFunc),
	}
}

// Start begins the schedule dispatcher loop
func (d *ScheduleDispatcher) Start(ctx context.Context) {
	d.mu.Lock()
	d.running = true
	d.mu.Unlock()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.Stop()
			return
		case <-ticker.C:
			d.dispatchDueTasks(ctx)
		}
	}
}

// Stop cancels all running scheduled tasks
func (d *ScheduleDispatcher) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.running = false
	for id, cancel := range d.tasks {
		cancel()
		delete(d.tasks, id)
	}
}

func (d *ScheduleDispatcher) dispatchDueTasks(ctx context.Context) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, name, type, cron_expr, config, status, run_count
		FROM scheduled_tasks
		WHERE status = 'active' AND (next_run_at IS NULL OR next_run_at <= NOW())
		LIMIT 10`)
	if err != nil {
		logger.Error("Failed to query scheduled tasks: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var task model.ScheduledTask
		if err := rows.Scan(&task.ID, &task.Name, &task.Type, &task.CronExpr, &task.Config, &task.Status, &task.RunCount); err != nil {
			continue
		}
		go d.executeTask(ctx, task)
	}
}

func (d *ScheduleDispatcher) executeTask(ctx context.Context, task model.ScheduledTask) {
	taskCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	d.mu.Lock()
	d.tasks[task.ID] = cancel
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		delete(d.tasks, task.ID)
		d.mu.Unlock()
	}()

	logger.Info("Executing scheduled task: %s (type=%s)", task.Name, task.Type)

	var taskErr error
	switch task.Type {
	case "metrics_aggregation":
		agg := NewMetricsAggregator()
		now := time.Now()
		taskErr = agg.AggregateHourly(taskCtx, "", now.Add(-time.Hour))
		if taskErr == nil {
			taskErr = agg.AggregateDaily(taskCtx, "", now.Add(-24*time.Hour))
		}
	case "data_quality":
		checker := NewDataQualityChecker()
		rules, err := fetchActiveQualityRules(taskCtx)
		if err == nil {
			for _, rule := range rules {
				checker.RunCheck(taskCtx, &rule)
			}
		}
	case "data_sync":
		engine := NewETLEngine()
		var config struct {
			Platform string `json:"platform"`
			Config   string `json:"config"`
		}
		json.Unmarshal([]byte(task.Config), &config)
		if config.Platform != "" {
			conn, err := CreateConnector(config.Platform, "", "")
			if err == nil {
				engine.RegisterConnector(config.Platform, conn)
				_, taskErr = engine.RunSync(taskCtx, config.Platform, config.Config)
			}
		}
	case "cache_warmup":
		agg := NewMetricsAggregator()
		taskErr = agg.WarmupCache(taskCtx)
	case "report_generation":
		// Will be implemented with report scheduler
		logger.Info("Report generation scheduled task: %s", task.Name)
	default:
		logger.Warn("Unknown scheduled task type: %s", task.Type)
	}

	// Update task status
	nextRun := calculateNextRun(task.CronExpr)
	status := "active"
	lastError := ""
	if taskErr != nil {
		lastError = taskErr.Error()
		if len(lastError) > 500 {
			lastError = lastError[:500]
		}
	}

	database.Get().Exec(taskCtx,
		`UPDATE scheduled_tasks SET last_run_at=NOW(), next_run_at=$1, run_count=run_count+1, last_error=$2, status=$3 WHERE id=$4`,
		nextRun, lastError, status, task.ID)
}

func fetchActiveQualityRules(ctx context.Context) ([]model.DataQualityRule, error) {
	rows, err := database.Get().Query(ctx,
		`SELECT id, name, table_name, column_name, rule_type, expression, severity, is_enabled, created_by
		FROM data_quality_rules WHERE is_enabled = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []model.DataQualityRule
	for rows.Next() {
		var r model.DataQualityRule
		if err := rows.Scan(&r.ID, &r.Name, &r.TableName, &r.ColumnName, &r.RuleType, &r.Expression, &r.Severity, &r.IsEnabled, &r.CreatedBy); err != nil {
			continue
		}
		rules = append(rules, r)
	}
	return rules, nil
}

func calculateNextRun(cronExpr string) time.Time {
	// Simplified: parse common patterns like "*/5 * * * *", "0 * * * *", etc.
	// Default to 1 hour from now
	return time.Now().Add(1 * time.Hour)
}
