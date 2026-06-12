-- ═══ Douyin Compass Session & Task Tables ═══
-- 抖音电商罗盘数据采集模块

-- Compass cookie sessions
CREATE TABLE IF NOT EXISTS compass_sessions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    cookie TEXT NOT NULL,                    -- encrypted browser cookies
    shop_id VARCHAR(50) NOT NULL,            -- 抖店 shop ID
    shop_name VARCHAR(200) DEFAULT '',
    user_agent TEXT DEFAULT '',               -- custom user-agent
    proxy_url VARCHAR(500) DEFAULT '',        -- optional proxy URL
    status VARCHAR(20) DEFAULT 'active',      -- active, expired, banned, cooldown, paused
    last_active_at TIMESTAMP,
    daily_requests INT DEFAULT 0,             -- today's request count
    max_daily_reqs INT DEFAULT 300,           -- daily limit (safe default)
    fail_count INT DEFAULT 0,                 -- consecutive failures
    next_available_at TIMESTAMP,              -- cooldown until this time
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_compass_sessions_status ON compass_sessions(status);
CREATE INDEX idx_compass_sessions_shop ON compass_sessions(shop_id);

-- Compass data collection tasks
CREATE TABLE IF NOT EXISTS compass_tasks (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES compass_sessions(id) ON DELETE CASCADE,
    task_type VARCHAR(30) NOT NULL,           -- live_overview, live_detail, product_list, product_detail, order_list, streamer_rank, funnel_analysis
    params JSONB DEFAULT '{}',                -- task parameters
    status VARCHAR(20) DEFAULT 'pending',     -- pending, running, completed, failed
    result JSONB,                             -- result summary
    records_count INT DEFAULT 0,
    error_msg TEXT DEFAULT '',
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_compass_tasks_session ON compass_tasks(session_id);
CREATE INDEX idx_compass_tasks_status ON compass_tasks(status);
CREATE INDEX idx_compass_tasks_type ON compass_tasks(task_type);
CREATE INDEX idx_compass_tasks_created ON compass_tasks(created_at DESC);

-- Collected live overview data
CREATE TABLE IF NOT EXISTS compass_live_data (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT REFERENCES compass_sessions(id),
    shop_id VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    live_count INT DEFAULT 0,
    total_duration INT DEFAULT 0,
    total_gmv DECIMAL(15,2) DEFAULT 0,
    total_orders INT DEFAULT 0,
    total_views BIGINT DEFAULT 0,
    total_likes BIGINT DEFAULT 0,
    total_comments BIGINT DEFAULT 0,
    avg_viewers INT DEFAULT 0,
    peak_viewers INT DEFAULT 0,
    new_followers INT DEFAULT 0,
    conversion_rate DECIMAL(5,4) DEFAULT 0,
    avg_order_value DECIMAL(10,2) DEFAULT 0,
    raw_data JSONB,
    collected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, date)
);

CREATE INDEX idx_compass_live_shop_date ON compass_live_data(shop_id, date DESC);

-- Collected product data
CREATE TABLE IF NOT EXISTS compass_product_data (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT REFERENCES compass_sessions(id),
    shop_id VARCHAR(50) NOT NULL,
    product_id VARCHAR(50) NOT NULL,
    title TEXT,
    category VARCHAR(100),
    price DECIMAL(10,2) DEFAULT 0,
    total_gmv DECIMAL(15,2) DEFAULT 0,
    total_orders INT DEFAULT 0,
    total_views BIGINT DEFAULT 0,
    conversion_rate DECIMAL(5,4) DEFAULT 0,
    raw_data JSONB,
    collected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_compass_product_shop ON compass_product_data(shop_id);
CREATE INDEX idx_compass_product_id ON compass_product_data(product_id);

-- Collected order data
CREATE TABLE IF NOT EXISTS compass_order_data (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT REFERENCES compass_sessions(id),
    shop_id VARCHAR(50) NOT NULL,
    order_id VARCHAR(50) NOT NULL,
    product_id VARCHAR(50),
    product_title TEXT,
    status VARCHAR(20),
    amount DECIMAL(10,2) DEFAULT 0,
    actual_amount DECIMAL(10,2) DEFAULT 0,
    quantity INT DEFAULT 1,
    commission DECIMAL(10,2) DEFAULT 0,
    live_room_id VARCHAR(50),
    ordered_at TIMESTAMP,
    raw_data JSONB,
    collected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, order_id)
);

CREATE INDEX idx_compass_order_shop ON compass_order_data(shop_id);
CREATE INDEX idx_compass_order_date ON compass_order_data(ordered_at DESC);
CREATE INDEX idx_compass_order_status ON compass_order_data(status);
