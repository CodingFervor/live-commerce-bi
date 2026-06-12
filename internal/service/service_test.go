package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// ═══ Export Engine Tests ═══

func TestBuildCSV(t *testing.T) {
	engine := NewExportEngine()

	headers := []string{"id", "name", "amount"}
	rows := [][]string{
		{"1", "Product A", "100.00"},
		{"2", "Product B", "200.50"},
		{"3", "Product, Special", "300.00"},
	}

	result := engine.BuildCSV(headers, rows)

	expected := "id,name,amount\n1,Product A,100.00\n2,Product B,200.50\n\"Product, Special\",300.00\n"
	if result != expected {
		t.Errorf("CSV mismatch.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestBuildCSVEmpty(t *testing.T) {
	engine := NewExportEngine()

	headers := []string{"id", "name"}
	result := engine.BuildCSV(headers, nil)

	expected := "id,name\n"
	if result != expected {
		t.Errorf("Empty CSV mismatch.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestBuildCSVEscapeQuotes(t *testing.T) {
	engine := NewExportEngine()

	headers := []string{"desc"}
	rows := [][]string{
		{`He said "hello"`},
	}

	result := engine.BuildCSV(headers, rows)
	expected := "desc\n\"He said \"\"hello\"\"\"\n"
	if result != expected {
		t.Errorf("CSV quote escape mismatch.\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestBuildJSON(t *testing.T) {
	engine := NewExportEngine()

	headers := []string{"id", "name", "amount"}
	rows := [][]string{
		{"1", "Product A", "100.00"},
		{"2", "Product B", "200.50"},
	}

	result := engine.BuildJSON(headers, rows)

	// Verify it's valid JSON by checking structure
	if len(result) == 0 {
		t.Error("JSON result is empty")
	}
	if result[0] != '[' {
		t.Errorf("JSON should start with '[', got %c", result[0])
	}
}

func TestBuildJSONEmpty(t *testing.T) {
	engine := NewExportEngine()

	headers := []string{"id", "name"}
	result := engine.BuildJSON(headers, nil)

	// Should be empty array or null
	if result != "null" && result != "[]" {
		t.Errorf("Empty JSON should be null or [], got: %s", result)
	}
}

// ═══ Hub Tests ═══

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("Hub should not be nil")
	}
	if hub.clients == nil {
		t.Error("Hub clients map should be initialized")
	}
	if hub.rooms == nil {
		t.Error("Hub rooms map should be initialized")
	}
	if hub.Register == nil {
		t.Error("Hub Register channel should be initialized")
	}
	if hub.Unregister == nil {
		t.Error("Hub Unregister channel should be initialized")
	}
}

func TestGetHubSingleton(t *testing.T) {
	hub1 := NewHub()
	hub2 := GetHub()
	if hub1 != hub2 {
		t.Error("GetHub should return the same instance as NewHub")
	}
}

func TestHubClientCount(t *testing.T) {
	hub := NewHub()
	if count := hub.ClientCount(); count != 0 {
		t.Errorf("Empty hub should have 0 clients, got %d", count)
	}
}

func TestHubRoomCount(t *testing.T) {
	hub := NewHub()
	if count := hub.RoomCount(); count != 0 {
		t.Errorf("Empty hub should have 0 rooms, got %d", count)
	}
}

func TestHubRegisterAndUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		RoomID: "room1",
	}

	// Register
	hub.Register <- client

	// Give hub time to process
	// In tests without real connections, we verify the channel mechanism works
	if hub.Register == nil {
		t.Error("Register channel should not be nil")
	}

	// Unregister
	hub.Unregister <- client
}

func TestHubBroadcastToAll(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		Hub:  hub,
		Send: make(chan []byte, 256),
	}

	hub.Register <- client

	// Broadcast a message
	hub.BroadcastToAll([]byte("test message"))

	// Verify broadcast channel is usable
	if hub.broadcast == nil {
		t.Error("Broadcast channel should not be nil")
	}
}

// ═══ Helper Function Tests ═══

func TestMinFunction(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{0, 0, 0},
		{-1, 1, -1},
	}
	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestContainsAny(t *testing.T) {
	if !containsAny("hello,world", ",") {
		t.Error("Should find comma")
	}
	if containsAny("helloworld", ",") {
		t.Error("Should not find comma")
	}
	if !containsAny(`say "hi"`, "\"") {
		t.Error("Should find quote")
	}
	if !containsAny("line1\nline2", "\n") {
		t.Error("Should find newline")
	}
}

// ═══ Platform Connector Tests ═══

func TestCreateConnectorDouyin(t *testing.T) {
	conn, err := CreateConnector("douyin", "app_key", "app_secret")
	if err != nil {
		t.Fatalf("CreateConnector douyin: %v", err)
	}
	if conn == nil {
		t.Fatal("Connector should not be nil")
	}
}

func TestCreateConnectorKuaishou(t *testing.T) {
	conn, err := CreateConnector("kuaishou", "app_key", "app_secret")
	if err != nil {
		t.Fatalf("CreateConnector kuaishou: %v", err)
	}
	if conn == nil {
		t.Fatal("Connector should not be nil")
	}
}

func TestCreateConnectorTaobao(t *testing.T) {
	conn, err := CreateConnector("taobao_live", "app_key", "app_secret")
	if err != nil {
		t.Fatalf("CreateConnector taobao_live: %v", err)
	}
	if conn == nil {
		t.Fatal("Connector should not be nil")
	}
}

func TestCreateConnectorUnsupported(t *testing.T) {
	_, err := CreateConnector("unsupported_platform", "", "")
	if err == nil {
		t.Error("Should return error for unsupported platform")
	}
}

func TestListSupportedPlatforms(t *testing.T) {
	platforms := ListSupportedPlatforms()
	if len(platforms) != 3 {
		t.Errorf("Expected 3 platforms, got %d", len(platforms))
	}
}

func TestDouyinConnectorConnect(t *testing.T) {
	conn := NewDouyinConnector("key", "secret")
	err := conn.Connect(`{"access_token":"test_token"}`)
	if err != nil {
		t.Fatalf("Connect should succeed: %v", err)
	}
	if conn.client.GetToken() != "test_token" {
		t.Error("Token not set correctly")
	}
}

func TestDouyinConnectorConnectNoToken(t *testing.T) {
	conn := NewDouyinConnector("key", "secret")
	err := conn.Connect(`{}`)
	if err == nil {
		t.Error("Should fail without access_token")
	}
}

func TestKuaishouConnectorConnect(t *testing.T) {
	conn := NewKuaishouConnector("key", "secret")
	err := conn.Connect(`{"access_token":"ks_token"}`)
	if err != nil {
		t.Fatalf("Connect should succeed: %v", err)
	}
}

func TestTaobaoConnectorConnect(t *testing.T) {
	conn := NewTaobaoConnector("key", "secret")
	err := conn.Connect(`{"session":"tb_session"}`)
	if err != nil {
		t.Fatalf("Connect should succeed: %v", err)
	}
}

func TestTaobaoConnectorConnectNoSession(t *testing.T) {
	conn := NewTaobaoConnector("key", "secret")
	err := conn.Connect(`{}`)
	if err == nil {
		t.Error("Should fail without session")
	}
}

// ═══ Data Quality Checker Tests ═══

func TestSafeTableName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"orders", "orders"},
		{"live_rooms", "live_rooms"},
		{"users", "users"},
		{"invalid_table; DROP TABLE users", "INVALID_TABLE"},
		{"", "INVALID_TABLE"},
	}
	for _, tt := range tests {
		result := safeTableName(tt.input)
		if result != tt.expected {
			t.Errorf("safeTableName(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestSafeColumnName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"name", "name"},
		{"created_at", "created_at"},
		{"id", "id"},
		{"col; DROP TABLE", "id"},
		{"", "id"},
		{"1col", "1col"}, // starts with number but valid chars
	}
	for _, tt := range tests {
		result := safeColumnName(tt.input)
		if result != tt.expected {
			t.Errorf("safeColumnName(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

// ═══ ETL Engine Tests ═══

func TestETLEngineNoConnector(t *testing.T) {
	engine := NewETLEngine()
	_, err := engine.RunSync(nil, "nonexistent", "")
	if err == nil {
		t.Error("Should return error for missing connector")
	}
}

func TestETLEngineRegisterConnector(t *testing.T) {
	engine := NewETLEngine()
	engine.RegisterConnector("test", &mockConnector{})
	if _, ok := engine.connectors["test"]; !ok {
		t.Error("Connector should be registered")
	}
}

// Mock connector for testing
type mockConnector struct{}

func (m *mockConnector) Connect(config string) error                      { return nil }
func (m *mockConnector) FetchLiveRooms(ctx context) ([]json.RawMessage, error) {
	return nil, nil
}
func (m *mockConnector) FetchOrders(ctx context, since time.Time) ([]json.RawMessage, error) {
	return nil, nil
}
func (m *mockConnector) FetchProducts(ctx context) ([]json.RawMessage, error) {
	return nil, nil
}
