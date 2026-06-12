-- Advanced SQL Schema - Part 1: Core tables
-- Run after init.sql

CREATE TABLE IF NOT EXISTS organizations (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    industry VARCHAR(50),
    plan VARCHAR(20) DEFAULT 'pro',
    max_users INT DEFAULT 50,
    max_rooms INT DEFAULT 100,
    status VARCHAR(20) DEFAULT 'active',
    expired_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS departments (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT REFERENCES organizations(id),
    parent_id BIGINT REFERENCES departments(id),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50),
    leader_id BIGINT,
    sort_order INT DEFAULT 0,
    path VARCHAR(500) DEFAULT '/',
    level INT DEFAULT 1,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_dept_org ON departments(organization_id);
CREATE INDEX IF NOT EXISTS idx_dept_parent ON departments(parent_id);

CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    module VARCHAR(30) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(20) NOT NULL
);

CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    organization_id BIGINT REFERENCES organizations(id),
    name VARCHAR(50) NOT NULL,
    code VARCHAR(50) NOT NULL,
    is_system BOOLEAN DEFAULT false,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, code)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT REFERENCES permissions(id) ON DELETE CASCADE,
    UNIQUE(role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT REFERENCES roles(id) ON DELETE CASCADE,
    UNIQUE(user_id, role_id)
);

CREATE TABLE IF NOT EXISTS data_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT REFERENCES roles(id) ON DELETE CASCADE,
    dimension VARCHAR(30) NOT NULL,
    values JSONB NOT NULL DEFAULT '[]',
    UNIQUE(role_id, dimension)
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    username VARCHAR(50),
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id VARCHAR(50),
    detail TEXT,
    ip VARCHAR(45),
    user_agent VARCHAR(500),
    request_id VARCHAR(50),
    duration INT DEFAULT 0,
    status_code INT DEFAULT 200,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_time ON audit_logs(created_at DESC);

CREATE TABLE IF NOT EXISTS track_events (
    id BIGSERIAL PRIMARY KEY,
    event_name VARCHAR(50) NOT NULL,
    platform VARCHAR(30),
    user_id VARCHAR(100),
    session_id VARCHAR(100),
    live_room_id BIGINT,
    product_id BIGINT,
    properties JSONB DEFAULT '{}',
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_evt_name ON track_events(event_name);
CREATE INDEX IF NOT EXISTS idx_evt_user ON track_events(user_id);
CREATE INDEX IF NOT EXISTS idx_evt_room ON track_events(live_room_id);
CREATE INDEX IF NOT EXISTS idx_evt_time ON track_events(timestamp DESC);

CREATE TABLE IF NOT EXISTS export_tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    type VARCHAR(30) NOT NULL,
    format VARCHAR(10) DEFAULT 'csv',
    params JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'pending',
    progress INT DEFAULT 0,
    total_rows INT DEFAULT 0,
    file_path VARCHAR(500),
    file_size BIGINT DEFAULT 0,
    error_msg TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS data_quality_rules (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    table_name VARCHAR(50) NOT NULL,
    column_name VARCHAR(50),
    rule_type VARCHAR(20) NOT NULL,
    expression TEXT,
    severity VARCHAR(20) DEFAULT 'warning',
    is_enabled BOOLEAN DEFAULT true,
    last_check_at TIMESTAMP,
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS data_quality_results (
    id BIGSERIAL PRIMARY KEY,
    rule_id BIGINT REFERENCES data_quality_rules(id),
    status VARCHAR(10) NOT NULL,
    total_rows INT DEFAULT 0,
    pass_rows INT DEFAULT 0,
    fail_rows INT DEFAULT 0,
    pass_rate DECIMAL(5,4) DEFAULT 0,
    detail JSONB DEFAULT '{}',
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(30) NOT NULL,
    cron_expr VARCHAR(50) NOT NULL,
    config JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active',
    last_run_at TIMESTAMP,
    next_run_at TIMESTAMP,
    run_count INT DEFAULT 0,
    last_error TEXT,
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notification_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    subject VARCHAR(200),
    content TEXT NOT NULL,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS notification_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    template_id BIGINT REFERENCES notification_templates(id),
    channel VARCHAR(20) NOT NULL,
    recipient VARCHAR(200) NOT NULL,
    subject VARCHAR(200),
    content TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    provider_id VARCHAR(100),
    retry_count INT DEFAULT 0,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS metrics_hourly (
    id BIGSERIAL PRIMARY KEY,
    platform VARCHAR(30) NOT NULL,
    hour TIMESTAMP NOT NULL,
    streamer_id BIGINT DEFAULT 0,
    live_room_id BIGINT DEFAULT 0,
    gmv DECIMAL(15,2) DEFAULT 0,
    order_count INT DEFAULT 0,
    viewer_count BIGINT DEFAULT 0,
    peak_viewers INT DEFAULT 0,
    like_count BIGINT DEFAULT 0,
    comment_count BIGINT DEFAULT 0,
    share_count BIGINT DEFAULT 0,
    conversion_rate DECIMAL(5,4) DEFAULT 0,
    avg_watch_time INT DEFAULT 0,
    new_followers INT DEFAULT 0,
    gift_count INT DEFAULT 0,
    gift_value DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(platform, hour, streamer_id, live_room_id)
);
CREATE INDEX IF NOT EXISTS idx_mh_hour ON metrics_hourly(platform, hour DESC);

CREATE TABLE IF NOT EXISTS metrics_daily (
    id BIGSERIAL PRIMARY KEY,
    platform VARCHAR(30) NOT NULL,
    date DATE NOT NULL,
    streamer_id BIGINT DEFAULT 0,
    live_room_count INT DEFAULT 0,
    total_duration INT DEFAULT 0,
    gmv DECIMAL(15,2) DEFAULT 0,
    order_count INT DEFAULT 0,
    viewer_count BIGINT DEFAULT 0,
    peak_viewers INT DEFAULT 0,
    like_count BIGINT DEFAULT 0,
    comment_count BIGINT DEFAULT 0,
    share_count BIGINT DEFAULT 0,
    conversion_rate DECIMAL(5,4) DEFAULT 0,
    avg_watch_time INT DEFAULT 0,
    new_followers INT DEFAULT 0,
    gift_count INT DEFAULT 0,
    gift_value DECIMAL(10,2) DEFAULT 0,
    refund_amount DECIMAL(15,2) DEFAULT 0,
    commission DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(platform, date, streamer_id)
);
CREATE INDEX IF NOT EXISTS idx_md_date ON metrics_daily(platform, date DESC);

CREATE TABLE IF NOT EXISTS dashboard_shares (
    id BIGSERIAL PRIMARY KEY,
    dashboard_id BIGINT REFERENCES dashboards(id) ON DELETE CASCADE,
    share_token VARCHAR(64) UNIQUE NOT NULL,
    password VARCHAR(50),
    allowed_roles JSONB DEFAULT '[]',
    expires_at TIMESTAMP,
    view_count INT DEFAULT 0,
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS live_room_timelines (
    id BIGSERIAL PRIMARY KEY,
    live_room_id BIGINT REFERENCES live_rooms(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    event_type VARCHAR(30) NOT NULL,
    title VARCHAR(200),
    data JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_timeline_room ON live_room_timelines(live_room_id);
CREATE INDEX IF NOT EXISTS idx_timeline_time ON live_room_timelines(timestamp DESC);

-- Seed permissions
INSERT INTO permissions (code, name, module, resource, action) VALUES
('dashboard:view', '查看仪表盘', 'dashboard', 'dashboard', 'view'),
('dashboard:create', '创建仪表盘', 'dashboard', 'dashboard', 'create'),
('dashboard:update', '编辑仪表盘', 'dashboard', 'dashboard', 'update'),
('dashboard:delete', '删除仪表盘', 'dashboard', 'dashboard', 'delete'),
('dashboard:share', '分享仪表盘', 'dashboard', 'dashboard', 'share'),
('analytics:view', '查看分析', 'analytics', 'analytics', 'view'),
('analytics:export', '导出分析数据', 'analytics', 'analytics', 'export'),
('liveroom:view', '查看直播间', 'liveroom', 'liveroom', 'view'),
('liveroom:manage', '管理直播间', 'liveroom', 'liveroom', 'manage'),
('streamer:view', '查看主播', 'streamer', 'streamer', 'view'),
('streamer:manage', '管理主播', 'streamer', 'streamer', 'manage'),
('product:view', '查看商品', 'product', 'product', 'view'),
('product:manage', '管理商品', 'product', 'product', 'manage'),
('order:view', '查看订单', 'order', 'order', 'view'),
('order:export', '导出订单', 'order', 'order', 'export'),
('report:view', '查看报表', 'report', 'report', 'view'),
('report:create', '创建报表', 'report', 'report', 'create'),
('report:export', '导出报表', 'report', 'report', 'export'),
('alert:view', '查看告警', 'alert', 'alert', 'view'),
('alert:manage', '管理告警', 'alert', 'alert', 'manage'),
('datasource:manage', '管理数据源', 'settings', 'datasource', 'manage'),
('user:manage', '管理用户', 'settings', 'user', 'manage'),
('role:manage', '管理角色', 'settings', 'role', 'manage')
ON CONFLICT (code) DO NOTHING;

-- Seed default organization and admin role
INSERT INTO organizations (name, code, industry, plan, max_users, max_rooms)
VALUES ('默认组织', 'default', '电商', 'enterprise', 500, 2000);

INSERT INTO roles (organization_id, name, code, is_system, description)
VALUES (1, '超级管理员', 'super_admin', true, '拥有所有权限');

INSERT INTO roles (organization_id, name, code, is_system, description)
VALUES (1, '数据分析师', 'analyst', true, '可查看分析和报表');

INSERT INTO roles (organization_id, name, code, is_system, description)
VALUES (1, '运营人员', 'operator', true, '可管理直播间和商品');

-- Assign all permissions to super_admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions;

-- Assign admin user to super_admin role
INSERT INTO user_roles (user_id, role_id) VALUES (1, 1);

-- Seed notification templates
INSERT INTO notification_templates (name, channel, subject, content, is_default) VALUES
('GMV预警', 'dingtalk', '直播GMV异常预警', '直播间 {{.room_title}} GMV {{.metric}} 触发预警，当前值: {{.value}}', true),
('订单日报', 'email', '直播电商订单日报', '今日GMV: {{.gmv}}, 订单数: {{.orders}}, 转化率: {{.conversion}}', true),
('退款预警', 'sms', '退款率过高', '退款率达到 {{.rate}}%，请及时处理', true);
