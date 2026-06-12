package repository

import (
	"context"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

// ═══ AI Config Repository ═══

type AIConfigRepo struct{}

func NewAIConfigRepo() *AIConfigRepo { return &AIConfigRepo{} }

func (r *AIConfigRepo) Create(ctx context.Context, cfg *model.AIConfig) error {
	return database.Get().QueryRow(ctx, `
		INSERT INTO ai_configs (name, provider, api_key, api_endpoint, model_name,
			max_tokens, temperature, top_p, is_default, is_enabled, proxy_url, extra_config, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, created_at`,
		cfg.Name, cfg.Provider, cfg.APIKey, cfg.APIEndpoint, cfg.ModelName,
		cfg.MaxTokens, cfg.Temperature, cfg.TopP, cfg.IsDefault, cfg.IsEnabled,
		cfg.ProxyURL, cfg.ExtraConfig, cfg.CreatedBy,
	).Scan(&cfg.ID, &cfg.CreatedAt)
}

func (r *AIConfigRepo) GetByID(ctx context.Context, id int64) (*model.AIConfig, error) {
	var cfg model.AIConfig
	err := database.Get().QueryRow(ctx, `
		SELECT id, name, provider, api_key, api_endpoint, model_name,
			max_tokens, temperature, top_p, is_default, is_enabled,
			proxy_url, extra_config, created_by, created_at, updated_at
		FROM ai_configs WHERE id=$1`, id).Scan(
		&cfg.ID, &cfg.Name, &cfg.Provider, &cfg.APIKey, &cfg.APIEndpoint,
		&cfg.ModelName, &cfg.MaxTokens, &cfg.Temperature, &cfg.TopP,
		&cfg.IsDefault, &cfg.IsEnabled, &cfg.ProxyURL, &cfg.ExtraConfig,
		&cfg.CreatedBy, &cfg.CreatedAt, &cfg.UpdatedAt)
	return &cfg, err
}

func (r *AIConfigRepo) List(ctx context.Context) ([]model.AIConfig, error) {
	rows, err := database.Get().Query(ctx, `
		SELECT id, name, provider, api_endpoint, model_name,
			max_tokens, temperature, top_p, is_default, is_enabled,
			created_by, created_at, updated_at
		FROM ai_configs ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.AIConfig
	for rows.Next() {
		var cfg model.AIConfig
		if err := rows.Scan(&cfg.ID, &cfg.Name, &cfg.Provider, &cfg.APIEndpoint,
			&cfg.ModelName, &cfg.MaxTokens, &cfg.Temperature, &cfg.TopP,
			&cfg.IsDefault, &cfg.IsEnabled, &cfg.CreatedBy, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
			continue
		}
		list = append(list, cfg)
	}
	return list, nil
}

func (r *AIConfigRepo) Update(ctx context.Context, cfg *model.AIConfig) error {
	_, err := database.Get().Exec(ctx, `
		UPDATE ai_configs SET name=$1, api_key=$2, api_endpoint=$3, model_name=$4,
			max_tokens=$5, temperature=$6, top_p=$7, is_default=$8, is_enabled=$9,
			proxy_url=$10, extra_config=$11, updated_at=NOW()
		WHERE id=$12`,
		cfg.Name, cfg.APIKey, cfg.APIEndpoint, cfg.ModelName,
		cfg.MaxTokens, cfg.Temperature, cfg.TopP, cfg.IsDefault, cfg.IsEnabled,
		cfg.ProxyURL, cfg.ExtraConfig, cfg.ID)
	return err
}

func (r *AIConfigRepo) Delete(ctx context.Context, id int64) error {
	_, err := database.Get().Exec(ctx, "DELETE FROM ai_configs WHERE id=$1", id)
	return err
}

func (r *AIConfigRepo) SetDefault(ctx context.Context, id int64) error {
	tx, err := database.Get().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tx.Exec(ctx, "UPDATE ai_configs SET is_default=false")
	tx.Exec(ctx, "UPDATE ai_configs SET is_default=true WHERE id=$1", id)
	return tx.Commit(ctx)
}

// ═══ AI Conversation Repository ═══

type AIConversationRepo struct{}

func NewAIConversationRepo() *AIConversationRepo { return &AIConversationRepo{} }

func (r *AIConversationRepo) Create(ctx context.Context, conv *model.AIConversation) error {
	return database.Get().QueryRow(ctx, `
		INSERT INTO ai_conversations (user_id, title, config_id, messages)
		VALUES ($1,$2,$3,$4) RETURNING id, created_at`,
		conv.UserID, conv.Title, conv.ConfigID, conv.Messages,
	).Scan(&conv.ID, &conv.CreatedAt)
}

func (r *AIConversationRepo) GetByID(ctx context.Context, id int64) (*model.AIConversation, error) {
	var conv model.AIConversation
	err := database.Get().QueryRow(ctx, `
		SELECT id, user_id, title, config_id, messages, created_at, updated_at
		FROM ai_conversations WHERE id=$1`, id).Scan(
		&conv.ID, &conv.UserID, &conv.Title, &conv.ConfigID,
		&conv.Messages, &conv.CreatedAt, &conv.UpdatedAt)
	return &conv, err
}

func (r *AIConversationRepo) ListByUser(ctx context.Context, userID int64, limit int) ([]model.AIConversation, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := database.Get().Query(ctx, `
		SELECT id, user_id, title, config_id, created_at, updated_at
		FROM ai_conversations WHERE user_id=$1 ORDER BY updated_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.AIConversation
	for rows.Next() {
		var c model.AIConversation
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.ConfigID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			continue
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *AIConversationRepo) Update(ctx context.Context, conv *model.AIConversation) error {
	_, err := database.Get().Exec(ctx, `
		UPDATE ai_conversations SET title=$1, messages=$2, updated_at=NOW() WHERE id=$3`,
		conv.Title, conv.Messages, conv.ID)
	return err
}

func (r *AIConversationRepo) Delete(ctx context.Context, id int64) error {
	_, err := database.Get().Exec(ctx, "DELETE FROM ai_conversations WHERE id=$1", id)
	return err
}

// ═══ System Settings Repository ═══

type SystemSettingRepo struct{}

func NewSystemSettingRepo() *SystemSettingRepo { return &SystemSettingRepo{} }

func (r *SystemSettingRepo) Get(ctx context.Context, category, key string) (*model.SystemSetting, error) {
	var s model.SystemSetting
	err := database.Get().QueryRow(ctx, `
		SELECT id, category, key, value, value_type, remark, is_public, created_at, updated_at
		FROM system_settings WHERE category=$1 AND key=$2`, category, key).Scan(
		&s.ID, &s.Category, &s.Key, &s.Value, &s.ValueType, &s.Remark, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

func (r *SystemSettingRepo) ListByCategory(ctx context.Context, category string) ([]model.SystemSetting, error) {
	rows, err := database.Get().Query(ctx, `
		SELECT id, category, key, value, value_type, remark, is_public, created_at, updated_at
		FROM system_settings WHERE category=$1 ORDER BY key`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.SystemSetting
	for rows.Next() {
		var s model.SystemSetting
		if err := rows.Scan(&s.ID, &s.Category, &s.Key, &s.Value, &s.ValueType, &s.Remark, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *SystemSettingRepo) ListAll(ctx context.Context) ([]model.SystemSetting, error) {
	rows, err := database.Get().Query(ctx, `
		SELECT id, category, key, value, value_type, remark, is_public, created_at, updated_at
		FROM system_settings ORDER BY category, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.SystemSetting
	for rows.Next() {
		var s model.SystemSetting
		if err := rows.Scan(&s.ID, &s.Category, &s.Key, &s.Value, &s.ValueType, &s.Remark, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *SystemSettingRepo) Upsert(ctx context.Context, s *model.SystemSetting) error {
	_, err := database.Get().Exec(ctx, `
		INSERT INTO system_settings (category, key, value, value_type, remark, is_public)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (category, key) DO UPDATE SET value=$3, remark=$5, updated_at=NOW()`,
		s.Category, s.Key, s.Value, s.ValueType, s.Remark, s.IsPublic)
	return err
}

func (r *SystemSettingRepo) BatchUpsert(ctx context.Context, settings []model.SystemSetting) error {
	tx, err := database.Get().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, s := range settings {
		_, err := tx.Exec(ctx, `
			INSERT INTO system_settings (category, key, value, value_type, remark, is_public)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (category, key) DO UPDATE SET value=$3, remark=$5, updated_at=NOW()`,
			s.Category, s.Key, s.Value, s.ValueType, s.Remark, s.IsPublic)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *SystemSettingRepo) Delete(ctx context.Context, category, key string) error {
	_, err := database.Get().Exec(ctx, "DELETE FROM system_settings WHERE category=$1 AND key=$2", category, key)
	return err
}
