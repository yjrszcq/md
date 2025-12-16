package service

import (
	"database/sql"
	"md/dao"
	"md/middleware"
	"md/model/common"
	"md/model/entity"
	"md/util"
	"time"
)

// 获取对话列表
func AIConversationList(userId string) []entity.AIConversationListItem {
	list, err := dao.AIConversationList(middleware.Db, userId)
	if err != nil {
		panic(common.NewErr("查询失败", err))
	}
	return list
}

// 搜索对话
func AIConversationSearch(userId string, keyword string) []entity.AIConversationListItem {
	if keyword == "" {
		return AIConversationList(userId)
	}
	list, err := dao.AIConversationSearch(middleware.Db, userId, keyword)
	if err != nil {
		panic(common.NewErr("搜索失败", err))
	}
	return list
}

// 获取对话详情
func AIConversationGet(userId string, id string) entity.AIConversation {
	conversation, err := dao.AIConversationGetById(middleware.Db, id, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			panic(common.NewErr("对话不存在", err))
		}
		panic(common.NewErr("查询失败", err))
	}
	return conversation
}

// 创建对话
func AIConversationAdd(userId string, condition entity.AIConversationCondition) string {
	tx := middleware.DbW.MustBegin()
	defer tx.Rollback()

	now := time.Now().UnixMilli()
	conversation := entity.AIConversation{
		Id:         util.SnowflakeString(),
		UserId:     userId,
		Title:      condition.Title,
		Content:    condition.Content,
		CreateTime: now,
		UpdateTime: now,
	}

	if conversation.Title == "" {
		conversation.Title = "新对话"
	}
	if conversation.Content == "" {
		conversation.Content = "[]"
	}

	err := dao.AIConversationAdd(tx, conversation)
	if err != nil {
		panic(common.NewErr("创建失败", err))
	}

	err = tx.Commit()
	if err != nil {
		panic(common.NewErr("创建失败", err))
	}

	return conversation.Id
}

// 更新对话
func AIConversationUpdate(userId string, condition entity.AIConversationCondition) {
	if condition.Id == "" {
		panic(common.NewError("对话ID不可为空"))
	}

	tx := middleware.DbW.MustBegin()
	defer tx.Rollback()

	now := time.Now().UnixMilli()
	conversation := entity.AIConversation{
		Id:         condition.Id,
		UserId:     userId,
		Title:      condition.Title,
		Content:    condition.Content,
		UpdateTime: now,
	}

	err := dao.AIConversationUpdate(tx, conversation)
	if err != nil {
		panic(common.NewErr("更新失败", err))
	}

	err = tx.Commit()
	if err != nil {
		panic(common.NewErr("更新失败", err))
	}
}

// 更新对话标题
func AIConversationUpdateTitle(userId string, id string, title string) {
	tx := middleware.DbW.MustBegin()
	defer tx.Rollback()

	now := time.Now().UnixMilli()
	err := dao.AIConversationUpdateTitle(tx, id, userId, title, now)
	if err != nil {
		panic(common.NewErr("更新失败", err))
	}

	err = tx.Commit()
	if err != nil {
		panic(common.NewErr("更新失败", err))
	}
}

// 删除对话
func AIConversationDelete(userId string, id string) {
	tx := middleware.DbW.MustBegin()
	defer tx.Rollback()

	err := dao.AIConversationDelete(tx, id, userId)
	if err != nil {
		panic(common.NewErr("删除失败", err))
	}

	err = tx.Commit()
	if err != nil {
		panic(common.NewErr("删除失败", err))
	}
}
