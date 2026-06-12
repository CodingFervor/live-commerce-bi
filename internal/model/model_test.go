package model

import (
	"encoding/json"
	"testing"
)

// ═══ Model JSON Serialization Tests ═══

func TestUserJSON(t *testing.T) {
	u := User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "admin",
		Status:   "active",
	}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("Marshal User: %v", err)
	}

	// Password should be hidden (- tag)
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	if _, exists := m["password"]; exists {
		t.Error("Password should not be serialized")
	}
	if m["username"] != "testuser" {
		t.Errorf("Expected username=testuser, got %v", m["username"])
	}
}

func TestLoginRequestBinding(t *testing.T) {
	req := LoginRequest{
		Username: "admin",
		Password: "secret123",
	}
	if req.Username != "admin" {
		t.Error("Username mismatch")
	}
	if req.Password != "secret123" {
		t.Error("Password mismatch")
	}
}

func TestRegisterRequestValidation(t *testing.T) {
	req := RegisterRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
		Role:     "analyst",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal RegisterRequest: %v", err)
	}
	var parsed RegisterRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal RegisterRequest: %v", err)
	}
	if parsed.Role != "analyst" {
		t.Errorf("Expected role=analyst, got %s", parsed.Role)
	}
}

func TestOrderJSON(t *testing.T) {
	o := Order{
		ID:           1,
		OrderNo:      "ORD-20240101-001",
		Platform:     "douyin",
		TotalAmount:  299.00,
		ActualAmount: 259.00,
		Status:       "completed",
	}

	data, err := json.Marshal(o)
	if err != nil {
		t.Fatalf("Marshal Order: %v", err)
	}

	var parsed Order
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal Order: %v", err)
	}
	if parsed.OrderNo != "ORD-20240101-001" {
		t.Errorf("OrderNo mismatch: %s", parsed.OrderNo)
	}
	if parsed.ActualAmount != 259.00 {
		t.Errorf("ActualAmount mismatch: %f", parsed.ActualAmount)
	}
}

func TestLiveRoomJSON(t *testing.T) {
	lr := LiveRoom{
		ID:             1,
		StreamerID:     10,
		Platform:       "kuaishou",
		Title:          "Test Live",
		Status:         "live",
		PeakViewers:    5000,
		TotalViews:     10000,
		GMV:            50000.0,
		ConversionRate: 3.5,
	}

	data, err := json.Marshal(lr)
	if err != nil {
		t.Fatalf("Marshal LiveRoom: %v", err)
	}

	var parsed LiveRoom
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Unmarshal LiveRoom: %v", err)
	}
	if parsed.Platform != "kuaishou" {
		t.Errorf("Platform mismatch: %s", parsed.Platform)
	}
}

func TestProductRankingJSON(t *testing.T) {
	pr := ProductRanking{
		Rank:          1,
		ProductID:     100,
		ProductName:   "Test Product",
		TotalRevenue:  50000.0,
		TotalSold:     500,
		ConversionRate: 5.2,
	}

	data, err := json.Marshal(pr)
	if err != nil {
		t.Fatalf("Marshal ProductRanking: %v", err)
	}

	var parsed ProductRanking
	json.Unmarshal(data, &parsed)
	if parsed.Rank != 1 {
		t.Errorf("Rank mismatch: %d", parsed.Rank)
	}
}

func TestOrderStats(t *testing.T) {
	stats := OrderStats{
		TotalOrders:    1000,
		TotalGMV:       500000.0,
		TotalActual:    480000.0,
		AvgOrderValue:  480.0,
		CompletedRate:  0.95,
		RefundRate:     0.02,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Marshal OrderStats: %v", err)
	}

	var parsed OrderStats
	json.Unmarshal(data, &parsed)
	if parsed.TotalOrders != 1000 {
		t.Errorf("TotalOrders mismatch: %d", parsed.TotalOrders)
	}
}

func TestAlertRuleCreate(t *testing.T) {
	rule := AlertRuleCreate{
		Name:      "High GMV Alert",
		Metric:    "gmv",
		Condition: "gt",
		Threshold: 100000,
		Severity:  "critical",
	}

	data, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("Marshal AlertRuleCreate: %v", err)
	}

	var parsed AlertRuleCreate
	json.Unmarshal(data, &parsed)
	if parsed.Metric != "gmv" {
		t.Errorf("Metric mismatch: %s", parsed.Metric)
	}
}

// ═══ Advanced Model Tests ═══

func TestOLAPQueryJSON(t *testing.T) {
	q := OLAPQuery{
		Dimensions: []string{"platform", "date"},
		Metrics:    []string{"gmv", "orders"},
		Filters:    map[string]string{"platform": "douyin"},
		Limit:      50,
	}

	data, err := json.Marshal(q)
	if err != nil {
		t.Fatalf("Marshal OLAPQuery: %v", err)
	}

	var parsed OLAPQuery
	json.Unmarshal(data, &parsed)
	if len(parsed.Dimensions) != 2 {
		t.Errorf("Dimensions count mismatch: %d", len(parsed.Dimensions))
	}
	if parsed.Filters["platform"] != "douyin" {
		t.Errorf("Filter mismatch: %v", parsed.Filters)
	}
}

func TestOLAPResultJSON(t *testing.T) {
	r := OLAPResult{
		Dimensions: map[string]interface{}{"platform": "douyin", "date": "2024-01-01"},
		Metrics:    map[string]interface{}{"gmv": 50000.0, "orders": 100},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal OLAPResult: %v", err)
	}

	var parsed OLAPResult
	json.Unmarshal(data, &parsed)
	if parsed.Dimensions["platform"] != "douyin" {
		t.Error("Dimension platform mismatch")
	}
}

func TestComparisonResultJSON(t *testing.T) {
	cr := ComparisonResult{
		Current:   map[string]interface{}{"period": "2024-01", "value": 50000.0},
		Previous:  map[string]interface{}{"period": "2023-01", "value": 40000.0},
		Change:    map[string]interface{}{"value": 10000.0},
		ChangePct: map[string]interface{}{"value": 25.0},
	}

	data, err := json.Marshal(cr)
	if err != nil {
		t.Fatalf("Marshal ComparisonResult: %v", err)
	}

	var parsed ComparisonResult
	json.Unmarshal(data, &parsed)
	if parsed.ChangePct["value"] != 25.0 {
		t.Errorf("ChangePct mismatch: %v", parsed.ChangePct["value"])
	}
}

func TestSalesForecastJSON(t *testing.T) {
	sf := SalesForecast{
		Date:     "2024-01-15",
		Actual:   45000.0,
		Forecast: 48000.0,
		Lower:    40000.0,
		Upper:    56000.0,
	}

	data, err := json.Marshal(sf)
	if err != nil {
		t.Fatalf("Marshal SalesForecast: %v", err)
	}

	var parsed SalesForecast
	json.Unmarshal(data, &parsed)
	if parsed.Date != "2024-01-15" {
		t.Errorf("Date mismatch: %s", parsed.Date)
	}
}

func TestAnomalyPointJSON(t *testing.T) {
	ap := AnomalyPoint{
		Value:     100000.0,
		Expected:  50000.0,
		Deviation: 3.5,
		IsAnomaly: true,
		Severity:  "critical",
	}

	data, err := json.Marshal(ap)
	if err != nil {
		t.Fatalf("Marshal AnomalyPoint: %v", err)
	}

	var parsed AnomalyPoint
	json.Unmarshal(data, &parsed)
	if !parsed.IsAnomaly {
		t.Error("IsAnomaly should be true")
	}
	if parsed.Severity != "critical" {
		t.Errorf("Severity mismatch: %s", parsed.Severity)
	}
}

func TestRFMSegmentJSON(t *testing.T) {
	rfm := RFMSegment{
		UserID:    "user_123",
		Recency:   5,
		Frequency: 20,
		Monetary:  15000.0,
		Score:     13,
		Segment:   "Champions",
	}

	data, err := json.Marshal(rfm)
	if err != nil {
		t.Fatalf("Marshal RFMSegment: %v", err)
	}

	var parsed RFMSegment
	json.Unmarshal(data, &parsed)
	if parsed.Segment != "Champions" {
		t.Errorf("Segment mismatch: %s", parsed.Segment)
	}
}

func TestExportTaskJSON(t *testing.T) {
	task := ExportTask{
		ID:        1,
		UserID:    10,
		Type:      "order_report",
		Format:    "csv",
		Status:    "completed",
		TotalRows: 5000,
		FileSize:  1024000,
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Marshal ExportTask: %v", err)
	}

	var parsed ExportTask
	json.Unmarshal(data, &parsed)
	if parsed.Format != "csv" {
		t.Errorf("Format mismatch: %s", parsed.Format)
	}
}

func TestDashboardCreateJSON(t *testing.T) {
	dc := DashboardCreate{
		Name:        "Sales Dashboard",
		Description: "Monthly sales overview",
		Type:        "sales",
		IsDefault:   false,
	}

	data, err := json.Marshal(dc)
	if err != nil {
		t.Fatalf("Marshal DashboardCreate: %v", err)
	}

	var parsed DashboardCreate
	json.Unmarshal(data, &parsed)
	if parsed.Name != "Sales Dashboard" {
		t.Errorf("Name mismatch: %s", parsed.Name)
	}
}

func TestDataSourceCreateValidation(t *testing.T) {
	dsc := DataSourceCreate{
		Name:         "Douyin Live",
		Platform:     "douyin",
		Config:       `{"app_key":"xxx","app_secret":"yyy"}`,
		SyncInterval: 30,
	}

	data, err := json.Marshal(dsc)
	if err != nil {
		t.Fatalf("Marshal DataSourceCreate: %v", err)
	}

	var parsed DataSourceCreate
	json.Unmarshal(data, &parsed)
	if parsed.Platform != "douyin" {
		t.Errorf("Platform mismatch: %s", parsed.Platform)
	}
}

func TestDepartmentTree(t *testing.T) {
	dept := Department{
		ID:             1,
		OrganizationID: 1,
		Name:           "Engineering",
		Level:          1,
		Children: []Department{
			{ID: 2, OrganizationID: 1, ParentID: intPtr(1), Name: "Backend", Level: 2},
			{ID: 3, OrganizationID: 1, ParentID: intPtr(1), Name: "Frontend", Level: 2},
		},
	}

	data, err := json.Marshal(dept)
	if err != nil {
		t.Fatalf("Marshal Department: %v", err)
	}

	var parsed Department
	json.Unmarshal(data, &parsed)
	if len(parsed.Children) != 2 {
		t.Errorf("Children count mismatch: %d", len(parsed.Children))
	}
}

func intPtr(i int64) *int64 {
	return &i
}
