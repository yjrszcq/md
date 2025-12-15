package service

import (
	"database/sql"
	"encoding/json"
	"md/dao"
	"md/middleware"
	"md/model/common"
	"md/model/entity"
	"md/util"
	"strings"
	"time"
)

// 获取AI配置（API Key脱敏）
func AIConfigGet(userId string) entity.AIConfigCondition {
	config, err := dao.AIConfigGetByUserId(middleware.Db, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.AIConfigCondition{SystemPrompts: []entity.SystemPrompt{}}
		}
		panic(common.NewErr("查询失败", err))
	}

	// 解析系统提示词
	var prompts []entity.SystemPrompt
	if config.SystemPrompts != "" && config.SystemPrompts != "[]" {
		json.Unmarshal([]byte(config.SystemPrompts), &prompts)
	}
	if prompts == nil {
		prompts = []entity.SystemPrompt{}
	}

	return entity.AIConfigCondition{
		BaseUrl:           config.BaseUrl,
		ApiKey:            maskApiKey(config.ApiKey),
		Model:             config.Model,
		SystemPrompts:     prompts,
		CurrentPromptId:   config.CurrentPromptId,
		AgentEnabled:      config.AgentEnabled,
		DocContextEnabled: config.DocContextEnabled,
		PanelEnabled:      config.PanelEnabled,
	}
}

// 保存AI配置
func AIConfigSave(userId string, condition entity.AIConfigCondition) {
	tx := middleware.DbW.MustBegin()
	defer tx.Rollback()

	// 检查是否存在配置
	exists, _ := dao.AIConfigExists(middleware.Db, userId)

	// 处理API Key
	apiKey := condition.ApiKey
	if apiKey != "" && !isMaskedKey(apiKey) {
		// 新的API Key，需要加密
		encrypted, err := util.EncryptAES(apiKey, common.AIEncryptKey, false)
		if err != nil {
			panic(common.NewErr("加密失败", err))
		}
		apiKey = encrypted
	} else if isMaskedKey(apiKey) {
		// 保持原有的API Key
		existing, err := dao.AIConfigGetByUserId(tx, userId)
		if err == nil {
			apiKey = existing.ApiKey
		} else {
			apiKey = ""
		}
	}

	// 序列化系统提示词
	promptsJson, _ := json.Marshal(condition.SystemPrompts)

	config := entity.AIConfig{
		UserId:            userId,
		BaseUrl:           strings.TrimSuffix(condition.BaseUrl, "/"),
		ApiKey:            apiKey,
		Model:             condition.Model,
		SystemPrompts:     string(promptsJson),
		CurrentPromptId:   condition.CurrentPromptId,
		AgentEnabled:      condition.AgentEnabled,
		DocContextEnabled: condition.DocContextEnabled,
		PanelEnabled:      condition.PanelEnabled,
		UpdateTime:        time.Now().UnixMilli(),
	}

	var err error
	if exists {
		err = dao.AIConfigUpdate(tx, config)
	} else {
		config.Id = util.SnowflakeString()
		config.CreateTime = time.Now().UnixMilli()
		err = dao.AIConfigAdd(tx, config)
	}

	if err != nil {
		panic(common.NewErr("保存失败", err))
	}

	err = tx.Commit()
	if err != nil {
		panic(common.NewErr("保存失败", err))
	}
}

// 删除AI配置
func AIConfigDelete(userId string) {
	tx := middleware.DbW.MustBegin()
	defer tx.Rollback()

	err := dao.AIConfigDelete(tx, userId)
	if err != nil {
		panic(common.NewErr("删除失败", err))
	}

	err = tx.Commit()
	if err != nil {
		panic(common.NewErr("删除失败", err))
	}
}

// 检查AI配置是否存在
func AIConfigExists(userId string) bool {
	exists, _ := dao.AIConfigExists(middleware.Db, userId)
	return exists
}

// 获取AI配置（完整版，包含解密后的API Key，仅用于同步）
func AIConfigGetFull(userId string) entity.AIConfigCondition {
	config, err := dao.AIConfigGetByUserId(middleware.Db, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.AIConfigCondition{SystemPrompts: []entity.SystemPrompt{}}
		}
		panic(common.NewErr("查询失败", err))
	}

	// 解析系统提示词
	var prompts []entity.SystemPrompt
	if config.SystemPrompts != "" && config.SystemPrompts != "[]" {
		json.Unmarshal([]byte(config.SystemPrompts), &prompts)
	}
	if prompts == nil {
		prompts = []entity.SystemPrompt{}
	}

	// 解密API Key
	apiKey := ""
	if config.ApiKey != "" {
		decrypted, err := util.DecryptAES(config.ApiKey, common.AIEncryptKey, false)
		if err == nil {
			apiKey = decrypted
		}
	}

	return entity.AIConfigCondition{
		BaseUrl:           config.BaseUrl,
		ApiKey:            apiKey,
		Model:             config.Model,
		SystemPrompts:     prompts,
		CurrentPromptId:   config.CurrentPromptId,
		AgentEnabled:      config.AgentEnabled,
		DocContextEnabled: config.DocContextEnabled,
		PanelEnabled:      config.PanelEnabled,
	}
}

// API Key脱敏
func maskApiKey(key string) string {
	if key == "" {
		return ""
	}
	// 解密后再脱敏
	decrypted, err := util.DecryptAES(key, common.AIEncryptKey, false)
	if err != nil {
		return "sk-..."
	}
	if len(decrypted) <= 8 {
		return "sk-..."
	}
	return decrypted[:7] + "..." + decrypted[len(decrypted)-4:]
}

// 判断是否为脱敏后的Key
func isMaskedKey(key string) bool {
	return key == "" || strings.Contains(key, "...")
}
