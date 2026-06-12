-- ═══ AI Configuration & System Settings Tables ═══
-- Live Commerce BI v2.1

-- AI Provider Configurations
CREATE TABLE IF NOT EXISTS ai_configs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    provider VARCHAR(30) NOT NULL, -- openai, qwen, zhipu, baidu, deepseek, moonshot, spark, ollama
    api_key TEXT NOT NULL DEFAULT '',
    api_endpoint VARCHAR(500) NOT NULL DEFAULT '',
    model_name VARCHAR(100) NOT NULL,
    max_tokens INT NOT NULL DEFAULT 4096,
    temperature DECIMAL(3,2) NOT NULL DEFAULT 0.70,
    top_p DECIMAL(3,2) NOT NULL DEFAULT 0.90,
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    proxy_url VARCHAR(500) NOT NULL DEFAULT '',
    extra_config TEXT DEFAULT '{}', -- provider-specific JSON
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_configs_provider ON ai_configs(provider);
CREATE INDEX idx_ai_configs_default ON ai_configs(is_default) WHERE is_default = true;

-- AI Chat Conversations
CREATE TABLE IF NOT EXISTS ai_conversations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    title VARCHAR(200) NOT NULL DEFAULT '',
    config_id BIGINT REFERENCES ai_configs(id),
    messages JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_conversations_user ON ai_conversations(user_id);
CREATE INDEX idx_ai_conversations_updated ON ai_conversations(updated_at DESC);

-- System Settings (Key-Value store)
CREATE TABLE IF NOT EXISTS system_settings (
    id BIGSERIAL PRIMARY KEY,
    category VARCHAR(50) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    value_type VARCHAR(20) NOT NULL DEFAULT 'string', -- string, int, float, bool, json
    remark VARCHAR(200) NOT NULL DEFAULT '',
    is_public BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(category, key)
);

CREATE INDEX idx_system_settings_category ON system_settings(category);

-- ═══ Seed: Default System Settings ═══

-- General
INSERT INTO system_settings (category, key, value, value_type, remark, is_public) VALUES
('general', 'site_name', '直播电商BI平台', 'string', '站点名称', true),
('general', 'site_logo', '', 'string', '站点Logo URL', true),
('general', 'default_language', 'zh-CN', 'string', '默认语言', true),
('general', 'timezone', 'Asia/Shanghai', 'string', '系统时区', true),
('general', 'date_format', 'YYYY-MM-DD', 'string', '日期格式', true),
('general', 'currency', 'CNY', 'string', '货币单位', true);

-- SMTP
INSERT INTO system_settings (category, key, value, value_type, remark, is_public) VALUES
('smtp', 'host', 'smtp.example.com', 'string', 'SMTP服务器地址', false),
('smtp', 'port', '465', 'int', 'SMTP端口', false),
('smtp', 'user', '', 'string', 'SMTP用户名', false),
('smtp', 'password', '', 'string', 'SMTP密码', false),
('smtp', 'from_name', '直播BI系统', 'string', '发件人名称', false),
('smtp', 'from_addr', 'bi@example.com', 'string', '发件人邮箱', false),
('smtp', 'use_tls', 'true', 'bool', '启用TLS', false);

-- Storage
INSERT INTO system_settings (category, key, value, value_type, remark, is_public) VALUES
('storage', 'provider', 'local', 'string', '存储类型(local/oss/s3/minio)', false),
('storage', 'endpoint', '', 'string', '存储端点', false),
('storage', 'bucket', '', 'string', '存储桶', false),
('storage', 'access_key', '', 'string', 'Access Key', false),
('storage', 'secret_key', '', 'string', 'Secret Key', false),
('storage', 'region', 'cn-hangzhou', 'string', '区域', false),
('storage', 'path_prefix', '/exports', 'string', '路径前缀', false);

-- Security
INSERT INTO system_settings (category, key, value, value_type, remark, is_public) VALUES
('security', 'password_min_length', '8', 'int', '密码最小长度', true),
('security', 'password_require_upper', 'false', 'bool', '密码需大写字母', true),
('security', 'password_require_number', 'true', 'bool', '密码需数字', true),
('security', 'password_require_special', 'false', 'bool', '密码需特殊字符', true),
('security', 'login_max_attempts', '5', 'int', '登录最大尝试次数', true),
('security', 'login_lock_duration', '30', 'int', '锁定时长(分钟)', true),
('security', 'session_timeout', '24', 'int', '会话超时(小时)', true),
('security', 'ip_whitelist', '', 'string', 'IP白名单(逗号分隔)', false),
('security', 'enable_2fa', 'false', 'bool', '启用双因素认证', true);

-- Notification
INSERT INTO system_settings (category, key, value, value_type, remark, is_public) VALUES
('notification', 'dingtalk_webhook', '', 'string', '钉钉机器人Webhook', false),
('notification', 'wechat_webhook', '', 'string', '企业微信Webhook', false),
('notification', 'sms_provider', '', 'string', '短信服务商', false),
('notification', 'sms_access_key', '', 'string', '短信AK', false),
('notification', 'sms_secret_key', '', 'string', '短信SK', false),
('notification', 'sms_sign_name', '直播BI', 'string', '短信签名', false);

-- AI Defaults
INSERT INTO system_settings (category, key, value, value_type, remark, is_public) VALUES
('ai', 'default_provider', 'openai', 'string', '默认AI提供商', true),
('ai', 'max_context_length', '10', 'int', '对话最大上下文轮数', true),
('ai', 'enable_nl_query', 'true', 'bool', '启用自然语言查询', true),
('ai', 'enable_auto_insight', 'true', 'bool', '启用自动洞察', true),
('ai', 'insight_schedule', '0 8 * * *', 'string', '洞察生成Cron', false),
('ai', 'system_prompt', '你是一个专业的直播电商BI数据分析师助手', 'string', '系统提示词', false);

-- ═══ Seed: Sample AI Config ═══

INSERT INTO ai_configs (name, provider, api_key, api_endpoint, model_name, max_tokens, temperature, top_p, is_default, is_enabled, created_by) VALUES
('OpenAI GPT-4o', 'openai', '', 'https://api.openai.com/v1/chat/completions', 'gpt-4o', 4096, 0.70, 0.90, true, true, 1),
('通义千问-Max', 'qwen', '', 'https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions', 'qwen-max', 4096, 0.70, 0.90, false, false, 1),
('DeepSeek-V3', 'deepseek', '', 'https://api.deepseek.com/v1/chat/completions', 'deepseek-chat', 4096, 0.70, 0.90, false, false, 1),
('智谱GLM-4', 'zhipu', '', 'https://open.bigmodel.cn/api/paas/v4/chat/completions', 'glm-4', 4096, 0.70, 0.90, false, false, 1),
('本地Ollama', 'ollama', '', 'http://localhost:11434/v1/chat/completions', 'qwen2.5:7b', 4096, 0.70, 0.90, false, false, 1);
