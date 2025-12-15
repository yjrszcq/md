package controller

import (
	"md/middleware"
	"md/model/common"
	"md/model/entity"
	"md/service"

	"github.com/kataras/iris/v12"
)

// 获取AI配置
func AIConfigGet(ctx iris.Context) {
	userId := middleware.CurrentUserId(ctx)
	ctx.JSON(common.NewSuccessData("查询成功", service.AIConfigGet(userId)))
}

// 保存AI配置
func AIConfigSave(ctx iris.Context) {
	condition := entity.AIConfigCondition{}
	resolveParam(ctx, &condition)
	userId := middleware.CurrentUserId(ctx)
	service.AIConfigSave(userId, condition)
	ctx.JSON(common.NewSuccess("保存成功"))
}

// 删除AI配置
func AIConfigDelete(ctx iris.Context) {
	userId := middleware.CurrentUserId(ctx)
	service.AIConfigDelete(userId)
	ctx.JSON(common.NewSuccess("删除成功"))
}

// 检查AI配置是否存在
func AIConfigExists(ctx iris.Context) {
	userId := middleware.CurrentUserId(ctx)
	exists := service.AIConfigExists(userId)
	ctx.JSON(common.NewSuccessData("查询成功", map[string]bool{"exists": exists}))
}

// 获取AI配置（完整版，用于同步）
func AIConfigGetFull(ctx iris.Context) {
	userId := middleware.CurrentUserId(ctx)
	ctx.JSON(common.NewSuccessData("查询成功", service.AIConfigGetFull(userId)))
}
