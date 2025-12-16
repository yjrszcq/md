package dao

import (
	"errors"
	"md/model/entity"

	"github.com/jmoiron/sqlx"
)

// 添加AI配置
func AIConfigAdd(tx *sqlx.Tx, config entity.AIConfig) error {
	sql := `insert into t_ai_config (id,user_id,base_url,api_key,model,system_prompts,current_prompt_id,system_prompt_enabled,doc_context_enabled,panel_enabled,create_time,update_time) values (:id,:user_id,:base_url,:api_key,:model,:system_prompts,:current_prompt_id,:system_prompt_enabled,:doc_context_enabled,:panel_enabled,:create_time,:update_time)`
	_, err := tx.NamedExec(sql, config)
	return err
}

// 更新AI配置
func AIConfigUpdate(tx *sqlx.Tx, config entity.AIConfig) error {
	sql := `update t_ai_config set base_url=:base_url,api_key=:api_key,model=:model,system_prompts=:system_prompts,current_prompt_id=:current_prompt_id,system_prompt_enabled=:system_prompt_enabled,doc_context_enabled=:doc_context_enabled,panel_enabled=:panel_enabled,update_time=:update_time where user_id=:user_id`
	_, err := tx.NamedExec(sql, config)
	return err
}

// 根据用户ID查询AI配置
func AIConfigGetByUserId(db interface{}, userId string) (entity.AIConfig, error) {
	sql := `select * from t_ai_config where user_id=$1`
	result := entity.AIConfig{}
	var err error
	switch db := db.(type) {
	case *sqlx.Tx:
		err = db.Get(&result, sql, userId)
	case *sqlx.DB:
		err = db.Get(&result, sql, userId)
	default:
		err = errors.New("数据库事务异常")
	}
	return result, err
}

// 删除AI配置
func AIConfigDelete(tx *sqlx.Tx, userId string) error {
	sql := `delete from t_ai_config where user_id=$1`
	_, err := tx.Exec(sql, userId)
	return err
}

// 检查AI配置是否存在
func AIConfigExists(db *sqlx.DB, userId string) (bool, error) {
	sql := `select count(*) from t_ai_config where user_id=$1`
	var count int
	err := db.Get(&count, sql, userId)
	return count > 0, err
}
