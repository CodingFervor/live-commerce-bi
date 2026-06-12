-- ═══════════════════════════════════════════════════════════════
-- Live Commerce BI - Enterprise BI System for Live-Streaming E-commerce
-- Database: live_commerce_bi
-- ═══════════════════════════════════════════════════════════════

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ─── Users ───
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'analyst' CHECK (role IN ('admin','analyst','viewer')),
    avatar VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active','inactive','banned')),
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- ─── Data Sources ───
CREATE TABLE data_sources (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    platform VARCHAR(30) NOT NULL CHECK (platform IN ('douyin','kuaishou','taobao_live','jd_live','pdd_live')),
    config JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active','inactive','error')),
    last_sync_at TIMESTAMP,
    sync_interval INT DEFAULT 300,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_data_sources_platform ON data_sources(platform);
CREATE INDEX idx_data_sources_status ON data_sources(status);

-- ─── Streamers / Anchors ───
CREATE TABLE streamers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    platform VARCHAR(30) NOT NULL,
    platform_id VARCHAR(100),
    avatar VARCHAR(255),
    follower_count BIGINT DEFAULT 0,
    category VARCHAR(50),
    tags TEXT[] DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active','inactive','banned')),
    data_source_id BIGINT REFERENCES data_sources(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_streamers_platform ON streamers(platform);
CREATE INDEX idx_streamers_category ON streamers(category);
CREATE INDEX idx_streamers_status ON streamers(status);
CREATE INDEX idx_streamers_data_source_id ON streamers(data_source_id);

-- ─── Live Rooms ───
CREATE TABLE live_rooms (
    id BIGSERIAL PRIMARY KEY,
    streamer_id BIGINT REFERENCES streamers(id),
    platform VARCHAR(30) NOT NULL,
    room_id VARCHAR(100),
    title VARCHAR(255),
    status VARCHAR(20) DEFAULT 'scheduled' CHECK (status IN ('scheduled','live','ended','cancelled')),
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    duration INT DEFAULT 0,
    peak_viewers INT DEFAULT 0,
    avg_viewers INT DEFAULT 0,
    total_views BIGINT DEFAULT 0,
    total_likes BIGINT DEFAULT 0,
    total_comments BIGINT DEFAULT 0,
    total_shares BIGINT DEFAULT 0,
    gmv DECIMAL(15,2) DEFAULT 0,
    order_count INT DEFAULT 0,
    product_count INT DEFAULT 0,
    conversion_rate DECIMAL(5,4) DEFAULT 0,
    tags TEXT[] DEFAULT '{}',
    data_source_id BIGINT REFERENCES data_sources(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_live_rooms_streamer_id ON live_rooms(streamer_id);
CREATE INDEX idx_live_rooms_platform ON live_rooms(platform);
CREATE INDEX idx_live_rooms_status ON live_rooms(status);
CREATE INDEX idx_live_rooms_started_at ON live_rooms(started_at);
CREATE INDEX idx_live_rooms_gmv ON live_rooms(gmv DESC);
CREATE INDEX idx_live_rooms_created_at ON live_rooms(created_at DESC);

-- ─── Products ───
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    platform VARCHAR(30) NOT NULL,
    platform_product_id VARCHAR(100),
    category VARCHAR(50),
    brand VARCHAR(100),
    price DECIMAL(10,2) DEFAULT 0,
    original_price DECIMAL(10,2) DEFAULT 0,
    live_price DECIMAL(10,2) DEFAULT 0,
    image_url VARCHAR(500),
    description TEXT,
    tags TEXT[] DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_products_platform ON products(platform);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_brand ON products(brand);
CREATE INDEX idx_products_status ON products(status);

-- ─── Live Room Products ───
CREATE TABLE live_room_products (
    id BIGSERIAL PRIMARY KEY,
    live_room_id BIGINT REFERENCES live_rooms(id) ON DELETE CASCADE,
    product_id BIGINT REFERENCES products(id),
    position INT DEFAULT 0,
    shown_at TIMESTAMP,
    highlighted BOOLEAN DEFAULT false,
    clicks INT DEFAULT 0,
    add_to_cart INT DEFAULT 0,
    orders INT DEFAULT 0,
    revenue DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_lrp_live_room_id ON live_room_products(live_room_id);
CREATE INDEX idx_lrp_product_id ON live_room_products(product_id);

-- ─── Orders ───
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(50) UNIQUE NOT NULL,
    platform VARCHAR(30) NOT NULL,
    platform_order_id VARCHAR(100),
    live_room_id BIGINT REFERENCES live_rooms(id),
    product_id BIGINT REFERENCES products(id),
    streamer_id BIGINT REFERENCES streamers(id),
    buyer_id VARCHAR(100),
    quantity INT DEFAULT 1,
    unit_price DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(10,2) DEFAULT 0,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    actual_amount DECIMAL(10,2) DEFAULT 0,
    commission_rate DECIMAL(5,4) DEFAULT 0,
    commission DECIMAL(10,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending','paid','shipped','completed','refunded','cancelled')),
    paid_at TIMESTAMP,
    shipped_at TIMESTAMP,
    completed_at TIMESTAMP,
    refunded_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_orders_order_no ON orders(order_no);
CREATE INDEX idx_orders_platform ON orders(platform);
CREATE INDEX idx_orders_live_room_id ON orders(live_room_id);
CREATE INDEX idx_orders_product_id ON orders(product_id);
CREATE INDEX idx_orders_streamer_id ON orders(streamer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_paid_at ON orders(paid_at);

-- ─── Viewer Metrics (time-series) ───
CREATE TABLE viewer_metrics (
    id BIGSERIAL PRIMARY KEY,
    live_room_id BIGINT REFERENCES live_rooms(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    concurrent_viewers INT DEFAULT 0,
    new_followers INT DEFAULT 0,
    likes INT DEFAULT 0,
    comments INT DEFAULT 0,
    shares INT DEFAULT 0,
    gifts_count INT DEFAULT 0,
    gifts_value DECIMAL(10,2) DEFAULT 0,
    avg_watch_time INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_vm_live_room_id ON viewer_metrics(live_room_id);
CREATE INDEX idx_vm_timestamp ON viewer_metrics(timestamp DESC);
CREATE INDEX idx_vm_live_room_timestamp ON viewer_metrics(live_room_id, timestamp DESC);

-- ─── Viewer Demographics ───
CREATE TABLE viewer_demographics (
    id BIGSERIAL PRIMARY KEY,
    live_room_id BIGINT REFERENCES live_rooms(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    age_distribution JSONB DEFAULT '{}',
    gender_distribution JSONB DEFAULT '{}',
    region_distribution JSONB DEFAULT '{}',
    device_distribution JSONB DEFAULT '{}',
    top_cities JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_vd_live_room_id ON viewer_demographics(live_room_id);
CREATE INDEX idx_vd_date ON viewer_demographics(date);

-- ─── Conversion Funnels ───
CREATE TABLE conversion_funnels (
    id BIGSERIAL PRIMARY KEY,
    live_room_id BIGINT REFERENCES live_rooms(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    impressions BIGINT DEFAULT 0,
    clicks BIGINT DEFAULT 0,
    add_to_carts BIGINT DEFAULT 0,
    orders BIGINT DEFAULT 0,
    payments BIGINT DEFAULT 0,
    click_rate DECIMAL(5,4) DEFAULT 0,
    cart_rate DECIMAL(5,4) DEFAULT 0,
    order_rate DECIMAL(5,4) DEFAULT 0,
    payment_rate DECIMAL(5,4) DEFAULT 0,
    overall_rate DECIMAL(5,4) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_cf_live_room_id ON conversion_funnels(live_room_id);
CREATE INDEX idx_cf_timestamp ON conversion_funnels(timestamp DESC);

-- ─── Dashboards ───
CREATE TABLE dashboards (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(30) DEFAULT 'custom' CHECK (type IN ('overview','realtime','custom')),
    layout JSONB DEFAULT '{}',
    is_default BOOLEAN DEFAULT false,
    owner_id BIGINT REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ─── Dashboard Widgets ───
CREATE TABLE dashboard_widgets (
    id BIGSERIAL PRIMARY KEY,
    dashboard_id BIGINT REFERENCES dashboards(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    type VARCHAR(30) NOT NULL CHECK (type IN ('line_chart','bar_chart','pie_chart','table','kpi_card','heatmap','funnel','map','gauge','scatter')),
    data_source VARCHAR(50),
    config JSONB DEFAULT '{}',
    position JSONB DEFAULT '{}',
    refresh_interval INT DEFAULT 30,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_dw_dashboard_id ON dashboard_widgets(dashboard_id);

-- ─── Report Templates ───
CREATE TABLE report_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(30) CHECK (type IN ('daily','weekly','monthly','custom')),
    sections JSONB DEFAULT '[]',
    thumbnail VARCHAR(255),
    is_public BOOLEAN DEFAULT true,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ─── Reports ───
CREATE TABLE reports (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    template_id BIGINT REFERENCES report_templates(id),
    type VARCHAR(30) NOT NULL CHECK (type IN ('daily','weekly','monthly','custom','one_time')),
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft','scheduled','generating','completed','failed')),
    config JSONB DEFAULT '{}',
    schedule VARCHAR(50),
    date_range_start DATE,
    date_range_end DATE,
    filters JSONB DEFAULT '{}',
    output_format VARCHAR(20) DEFAULT 'pdf' CHECK (output_format IN ('pdf','excel','html','csv')),
    file_path VARCHAR(500),
    generated_at TIMESTAMP,
    generated_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_reports_status ON reports(status);
CREATE INDEX idx_reports_type ON reports(type);
CREATE INDEX idx_reports_created_at ON reports(created_at DESC);

-- ─── Alert Rules ───
CREATE TABLE alert_rules (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    metric VARCHAR(50) NOT NULL,
    condition VARCHAR(20) NOT NULL CHECK (condition IN ('gt','lt','eq','gte','lte','change_pct_gt','change_pct_lt')),
    threshold DECIMAL(15,4) NOT NULL,
    timeframe VARCHAR(20) DEFAULT '5m',
    severity VARCHAR(20) DEFAULT 'warning' CHECK (severity IN ('info','warning','critical')),
    notify_channels JSONB DEFAULT '[]',
    notify_config JSONB DEFAULT '{}',
    cooldown INT DEFAULT 300,
    is_enabled BOOLEAN DEFAULT true,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ─── Alert History ───
CREATE TABLE alert_history (
    id BIGSERIAL PRIMARY KEY,
    alert_rule_id BIGINT REFERENCES alert_rules(id),
    metric VARCHAR(50),
    triggered_value DECIMAL(15,4),
    threshold_value DECIMAL(15,4),
    message TEXT,
    severity VARCHAR(20),
    notified BOOLEAN DEFAULT false,
    acknowledged BOOLEAN DEFAULT false,
    acknowledged_by BIGINT REFERENCES users(id),
    acknowledged_at TIMESTAMP,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_ah_alert_rule_id ON alert_history(alert_rule_id);
CREATE INDEX idx_ah_severity ON alert_history(severity);
CREATE INDEX idx_ah_acknowledged ON alert_history(acknowledged);
CREATE INDEX idx_ah_created_at ON alert_history(created_at DESC);

-- ─── Revenue Records ───
CREATE TABLE revenue_records (
    id BIGSERIAL PRIMARY KEY,
    date DATE NOT NULL,
    platform VARCHAR(30) NOT NULL,
    live_room_id BIGINT REFERENCES live_rooms(id),
    streamer_id BIGINT REFERENCES streamers(id),
    gmv DECIMAL(15,2) DEFAULT 0,
    actual_revenue DECIMAL(15,2) DEFAULT 0,
    commission DECIMAL(15,2) DEFAULT 0,
    refund_amount DECIMAL(15,2) DEFAULT 0,
    order_count INT DEFAULT 0,
    avg_order_value DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(date, platform, live_room_id)
);
CREATE INDEX idx_rr_date ON revenue_records(date DESC);
CREATE INDEX idx_rr_platform ON revenue_records(platform);
CREATE INDEX idx_rr_streamer_id ON revenue_records(streamer_id);

-- ─── Platform Metrics (hourly) ───
CREATE TABLE platform_metrics (
    id BIGSERIAL PRIMARY KEY,
    platform VARCHAR(30) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    live_room_count INT DEFAULT 0,
    total_viewers BIGINT DEFAULT 0,
    total_gmv DECIMAL(15,2) DEFAULT 0,
    total_orders INT DEFAULT 0,
    avg_conversion_rate DECIMAL(5,4) DEFAULT 0,
    top_category VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(platform, timestamp)
);
CREATE INDEX idx_pm_platform ON platform_metrics(platform);
CREATE INDEX idx_pm_timestamp ON platform_metrics(timestamp DESC);

-- ─── Inventory Snapshots ───
CREATE TABLE inventory_snapshots (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT REFERENCES products(id),
    live_room_id BIGINT REFERENCES live_rooms(id),
    timestamp TIMESTAMP NOT NULL,
    stock_before INT,
    stock_after INT,
    sold INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_is_product_id ON inventory_snapshots(product_id);
CREATE INDEX idx_is_timestamp ON inventory_snapshots(timestamp DESC);

-- ─── Sync Logs ───
CREATE TABLE sync_logs (
    id BIGSERIAL PRIMARY KEY,
    data_source_id BIGINT REFERENCES data_sources(id),
    sync_type VARCHAR(30) CHECK (sync_type IN ('full','incremental','realtime')),
    status VARCHAR(20) CHECK (status IN ('running','completed','failed')),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    records_processed INT DEFAULT 0,
    records_failed INT DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_sl_data_source_id ON sync_logs(data_source_id);
CREATE INDEX idx_sl_status ON sync_logs(status);
CREATE INDEX idx_sl_created_at ON sync_logs(created_at DESC);

-- ─── Seed: Admin User (password: admin123) ───
INSERT INTO users (username, email, password, role, status)
VALUES ('admin', 'admin@livecommerce.bi', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin', 'active');

-- ─── Seed: Default Dashboard ───
INSERT INTO dashboards (name, description, type, is_default, status)
VALUES ('实时直播监控', '实时监控所有正在直播的直播间数据', 'realtime', true, 'active');

INSERT INTO dashboards (name, description, type, is_default, status)
VALUES ('销售总览', 'GMV、订单、转化率等核心指标概览', 'overview', true, 'active');

-- ─── Seed: Report Templates ───
INSERT INTO report_templates (name, description, type, sections, is_public)
VALUES ('日报模板', '每日直播电商数据汇总', 'daily', '[{"title":"GMV概览","type":"kpi"},{"title":"订单分析","type":"table"},{"title":"主播排名","type":"bar_chart"},{"title":"商品TOP10","type":"bar_chart"}]', true);

INSERT INTO report_templates (name, description, type, sections, is_public)
VALUES ('周报模板', '每周直播电商趋势分析', 'weekly', '[{"title":"周度GMV趋势","type":"line_chart"},{"title":"平台对比","type":"pie_chart"},{"title":"主播排行","type":"table"},{"title":"转化漏斗","type":"funnel"},{"title":"商品分析","type":"table"}]', true);

INSERT INTO report_templates (name, description, type, sections, is_public)
VALUES ('月报模板', '月度直播电商综合分析', 'monthly', '[{"title":"月度概览","type":"kpi"},{"title":"GMV趋势","type":"line_chart"},{"title":"品类分析","type":"pie_chart"},{"title":"主播绩效","type":"table"},{"title":"用户画像","type":"pie_chart"},{"title":"转化分析","type":"funnel"}]', true);

-- ─── Seed: Sample Alert Rules ───
INSERT INTO alert_rules (name, description, metric, condition, threshold, timeframe, severity, notify_channels, is_enabled)
VALUES ('GMV骤降预警', '当5分钟内GMV下降超过30%时触发', 'gmv', 'change_pct_lt', -30.0, '5m', 'critical', '["email","dingtalk"]', true);

INSERT INTO alert_rules (name, description, metric, condition, threshold, timeframe, severity, notify_channels, is_enabled)
VALUES ('观众人数异常', '当并发观众数低于50时触发', 'viewers', 'lt', 50.0, '1m', 'warning', '["dingtalk"]', true);

INSERT INTO alert_rules (name, description, metric, condition, threshold, timeframe, severity, notify_channels, is_enabled)
VALUES ('退款率过高', '当退款率超过15%时触发', 'refund_rate', 'gt', 15.0, '1h', 'critical', '["email","sms"]', true);

-- ─── Views ───
CREATE OR REPLACE VIEW v_live_room_stats AS
SELECT
    lr.id,
    lr.title,
    lr.platform,
    lr.status,
    lr.started_at,
    lr.ended_at,
    lr.duration,
    lr.peak_viewers,
    lr.avg_viewers,
    lr.total_views,
    lr.gmv,
    lr.order_count,
    lr.conversion_rate,
    s.name AS streamer_name,
    s.category,
    COUNT(DISTINCT lrp.product_id) AS product_count,
    COALESCE(SUM(o.actual_amount), 0) AS actual_revenue,
    COALESCE(SUM(CASE WHEN o.status = 'completed' THEN o.commission ELSE 0 END), 0) AS total_commission
FROM live_rooms lr
JOIN streamers s ON lr.streamer_id = s.id
LEFT JOIN live_room_products lrp ON lr.id = lrp.live_room_id
LEFT JOIN orders o ON lr.id = o.live_room_id
GROUP BY lr.id, s.name, s.category;

CREATE OR REPLACE VIEW v_streamer_summary AS
SELECT
    s.id,
    s.name,
    s.platform,
    s.category,
    s.follower_count,
    COUNT(DISTINCT lr.id) AS total_live_rooms,
    COALESCE(SUM(lr.gmv), 0) AS total_gmv,
    COALESCE(SUM(lr.order_count), 0) AS total_orders,
    COALESCE(SUM(lr.total_views), 0) AS total_views,
    COALESCE(AVG(lr.conversion_rate), 0) AS avg_conversion,
    COALESCE(AVG(lr.avg_viewers), 0) AS avg_viewers,
    COALESCE(SUM(lr.duration), 0) AS total_duration
FROM streamers s
LEFT JOIN live_rooms lr ON s.id = lr.streamer_id
GROUP BY s.id;

CREATE OR REPLACE VIEW v_product_summary AS
SELECT
    p.id,
    p.name,
    p.brand,
    p.category,
    p.price,
    p.platform,
    COUNT(DISTINCT lrp.live_room_id) AS live_room_count,
    COALESCE(SUM(lrp.orders), 0) AS total_sold,
    COALESCE(SUM(lrp.revenue), 0) AS total_revenue,
    COALESCE(AVG(CASE WHEN lrp.clicks > 0 THEN lrp.orders::FLOAT / lrp.clicks ELSE 0 END), 0) AS avg_conversion
FROM products p
LEFT JOIN live_room_products lrp ON p.id = lrp.product_id
GROUP BY p.id;

-- ─── Functions ───
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at trigger to key tables
CREATE TRIGGER trg_users_updated BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_data_sources_updated BEFORE UPDATE ON data_sources FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_streamers_updated BEFORE UPDATE ON streamers FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_live_rooms_updated BEFORE UPDATE ON live_rooms FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_products_updated BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_orders_updated BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_dashboards_updated BEFORE UPDATE ON dashboards FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_reports_updated BEFORE UPDATE ON reports FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_alert_rules_updated BEFORE UPDATE ON alert_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER trg_revenue_records_updated BEFORE UPDATE ON revenue_records FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ─── Materialized View for Dashboard Performance ───
CREATE MATERIALIZED VIEW mv_daily_gmv AS
SELECT
    DATE(o.created_at) AS date,
    o.platform,
    COUNT(DISTINCT o.id) AS order_count,
    COALESCE(SUM(o.total_amount), 0) AS gmv,
    COALESCE(SUM(o.actual_amount), 0) AS actual_revenue,
    COALESCE(SUM(o.commission), 0) AS commission,
    COALESCE(SUM(CASE WHEN o.status = 'refunded' THEN o.actual_amount ELSE 0 END), 0) AS refund_amount,
    COALESCE(AVG(o.actual_amount), 0) AS avg_order_value
FROM orders o
WHERE o.status NOT IN ('pending', 'cancelled')
GROUP BY DATE(o.created_at), o.platform
ORDER BY date DESC;

CREATE UNIQUE INDEX idx_mv_daily_gmv_date_platform ON mv_daily_gmv(date, platform);

REFRESH MATERIALIZED VIEW mv_daily_gmv;

-- ─── Performance Indexes ───
-- Missing indexes identified by performance audit

-- orders.buyer_id: used by cohort analysis, RFM analysis, and buyer lookups
CREATE INDEX IF NOT EXISTS idx_orders_buyer_id ON orders(buyer_id);
CREATE INDEX IF NOT EXISTS idx_orders_buyer_status ON orders(buyer_id, status);

-- dashboards.owner_id: used by ListDashboards (per-user query)
CREATE INDEX IF NOT EXISTS idx_dashboards_owner_id ON dashboards(owner_id);

-- alert_rules composite: common query pattern for alert engine
CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(is_enabled) WHERE is_enabled = true;

-- export_tasks: list by user + status
CREATE INDEX IF NOT EXISTS idx_export_tasks_user_status ON export_tasks(user_id, status);

-- track_events composite: funnel and path analysis queries
CREATE INDEX IF NOT EXISTS idx_evt_room_time ON track_events(live_room_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_evt_session ON track_events(session_id);

-- live_rooms composite: analytics dashboard queries by platform + date
CREATE INDEX IF NOT EXISTS idx_lr_platform_created ON live_rooms(platform, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_lr_status_created ON live_rooms(status, created_at DESC);

-- Covering index for dashboard overview stats
CREATE INDEX IF NOT EXISTS idx_lr_gmv_date ON live_rooms(created_at DESC, platform, gmv, order_count, total_views, conversion_rate);
