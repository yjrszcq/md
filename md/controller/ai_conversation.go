package controller

import (
	"md/middleware"
	"md/model/common"
	"md/model/entity"
	"md/service"

	"github.com/kataras/iris/v12"
)

// 获取对话列表
func AIConversationList(ctx iris.Context) {
	userId := middleware.CurrentUserId(ctx)
	ctx.JSON(common.NewSuccessData("查询成功", service.AIConversationList(userId)))
}

// 搜索对话
func AIConversationSearch(ctx iris.Context) {
	condition := entity.AIConversationCondition{}
	resolveParam(ctx, &condition)
	userId := middleware.CurrentUserId(ctx)
	ctx.JSON(common.NewSuccessData("查询成功", service.AIConversationSearch(userId, condition.Keyword)))
}

// 获取对话详情
func AIConversationGet(ctx iris.Context) {
	condition := entity.AIConversationCondition{}
	resolveParam(ctx, &condition)
	userId := middleware.CurrentUserId(ctx)
	ctx.JSON(common.NewSuccessData("查询成功", service.AIConversationGet(userId, condition.Id)))
}

// 创建对话
func AIConversationAdd(ctx iris.Context) {
	condition := entity.AIConversationCondition{}
	resolveParam(ctx, &condition)
	userId := middleware.CurrentUserId(ctx)
	id := service.AIConversationAdd(userId, condition)
	ctx.JSON(common.NewSuccessData("创建成功", map[string]string{"id": id}))
}

// 更新对话
func AIConversationUpdate(ctx iris.Context) {
	condition := entity.AIConversationCondition{}
	resolveParam(ctx, &condition)
	userId := middleware.CurrentUserId(ctx)
	service.AIConversationUpdate(userId, condition)
	ctx.JSON(common.NewSuccess("更新成功"))
}

// 更新对话标题
func AIConversationUpdateTitle(ctx iris.Context) {
	condition := entity.AIConversationCondition{}
	resolveParam(ctx, &condition)
	userId := middleware.CurrentUserId(ctx)
	service.AIConversationUpdateTitle(userId, condition.Id, condition.Title)
	ctx.JSON(common.NewSuccess("更新成功"))
}

// 删除对话
func AIConversationDelete(ctx iris.Context) {
	condition := entity.AIConversationCondition{}
	resolveParam(ctx, &condition)
	userId := middleware.CurrentUserId(ctx)
	service.AIConversationDelete(userId, condition.Id)
	ctx.JSON(common.NewSuccess("删除成功"))
}
