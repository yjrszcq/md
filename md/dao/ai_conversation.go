package dao

import (
	"errors"
	"md/model/entity"

	"github.com/jmoiron/sqlx"
)

// 添加对话
func AIConversationAdd(tx *sqlx.Tx, conversation entity.AIConversation) error {
	sql := `insert into t_ai_conversation (id,user_id,title,content,create_time,update_time) values (:id,:user_id,:title,:content,:create_time,:update_time)`
	_, err := tx.NamedExec(sql, conversation)
	return err
}

// 更新对话
func AIConversationUpdate(tx *sqlx.Tx, conversation entity.AIConversation) error {
	sql := `update t_ai_conversation set title=:title,content=:content,update_time=:update_time where id=:id and user_id=:user_id`
	_, err := tx.NamedExec(sql, conversation)
	return err
}

// 更新对话标题
func AIConversationUpdateTitle(tx *sqlx.Tx, id string, userId string, title string, updateTime int64) error {
	sql := `update t_ai_conversation set title=$1,update_time=$2 where id=$3 and user_id=$4`
	_, err := tx.Exec(sql, title, updateTime, id, userId)
	return err
}

// 根据ID查询对话
func AIConversationGetById(db interface{}, id string, userId string) (entity.AIConversation, error) {
	sql := `select * from t_ai_conversation where id=$1 and user_id=$2`
	result := entity.AIConversation{}
	var err error
	switch db := db.(type) {
	case *sqlx.Tx:
		err = db.Get(&result, sql, id, userId)
	case *sqlx.DB:
		err = db.Get(&result, sql, id, userId)
	default:
		err = errors.New("数据库事务异常")
	}
	return result, err
}

// 查询用户的对话列表（不含内容）
func AIConversationList(db *sqlx.DB, userId string) ([]entity.AIConversationListItem, error) {
	sql := `select id,title,create_time,update_time from t_ai_conversation where user_id=$1 order by update_time desc`
	result := []entity.AIConversationListItem{}
	err := db.Select(&result, sql, userId)
	return result, err
}

// 搜索对话（标题和内容）
func AIConversationSearch(db *sqlx.DB, userId string, keyword string) ([]entity.AIConversationListItem, error) {
	sql := `select id,title,create_time,update_time from t_ai_conversation where user_id=$1 and (title like $2 or content like $2) order by update_time desc`
	result := []entity.AIConversationListItem{}
	err := db.Select(&result, sql, userId, "%"+keyword+"%")
	return result, err
}

// 删除对话
func AIConversationDelete(tx *sqlx.Tx, id string, userId string) error {
	sql := `delete from t_ai_conversation where id=$1 and user_id=$2`
	_, err := tx.Exec(sql, id, userId)
	return err
}

// 删除用户的所有对话
func AIConversationDeleteByUserId(tx *sqlx.Tx, userId string) error {
	sql := `delete from t_ai_conversation where user_id=$1`
	_, err := tx.Exec(sql, userId)
	return err
}
