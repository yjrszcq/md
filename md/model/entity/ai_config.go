package entity

// AIConfig AI配置实体
type AIConfig struct {
	Id                  string `json:"id" db:"id"`
	UserId              string `json:"userId" db:"user_id"`
	BaseUrl             string `json:"baseUrl" db:"base_url"`
	ApiKey              string `json:"apiKey" db:"api_key"` // 加密存储
	Model               string `json:"model" db:"model"`
	SystemPrompts       string `json:"systemPrompts" db:"system_prompts"` // JSON数组
	CurrentPromptId     string `json:"currentPromptId" db:"current_prompt_id"`
	SystemPromptEnabled bool   `json:"systemPromptEnabled" db:"system_prompt_enabled"`
	DocContextEnabled   bool   `json:"docContextEnabled" db:"doc_context_enabled"`
	PanelEnabled        bool   `json:"panelEnabled" db:"panel_enabled"`
	CreateTime          int64  `json:"createTime" db:"create_time"`
	UpdateTime          int64  `json:"updateTime" db:"update_time"`
}

// SystemPrompt 系统提示词
type SystemPrompt struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	IsActive bool   `json:"isActive"`
}

// AIConfigCondition AI配置请求/响应条件
type AIConfigCondition struct {
	BaseUrl             string         `json:"baseUrl"`
	ApiKey              string         `json:"apiKey"`
	Model               string         `json:"model"`
	SystemPrompts       []SystemPrompt `json:"systemPrompts"`
	CurrentPromptId     string         `json:"currentPromptId"`
	SystemPromptEnabled bool           `json:"systemPromptEnabled"`
	DocContextEnabled   bool           `json:"docContextEnabled"`
	PanelEnabled        bool           `json:"panelEnabled"`
}
