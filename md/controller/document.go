package controller

import (
	"md/middleware"
	"md/model/common"
	"md/model/entity"
	"md/service"

	"github.com/kataras/iris/v12"
)

// 添加文档
func DocumentAdd(ctx iris.Context) {
	document := entity.Document{}
	resolveParam(ctx, &document)
	document.UserId = middleware.CurrentUserId(ctx)
	ctx.JSON(common.NewSuccessData("添加成功", service.DocumentAdd(document)))
}

// 修改文档基础信息
func DocumentUpdate(ctx iris.Context) {
	document := entity.Document{}
	resolveParam(ctx, &document)
	document.UserId = middleware.CurrentUserId(ctx)
	service.DocumentUpdate(document)
	ctx.JSON(common.NewSuccess("更新成功"))
}

// 修改文档内容
func DocumentUpdateContent(ctx iris.Context) {
	document := entity.Document{}
	resolveParam(ctx, &document)
	document.UserId = middleware.CurrentUserId(ctx)
	ctx.JSON(common.NewSuccessData("更新成功", service.DocumentUpdateContent(document)))
}

// 删除文档
func DocumentDelete(ctx iris.Context) {
	document := entity.Document{}
	resolveParam(ctx, &document)
	userId := middleware.CurrentUserId(ctx)
	service.DocumentDelete(document.Id, userId)
	ctx.JSON(common.NewSuccess("删除成功"))
}

// 查询文档列表
func DocumentList(ctx iris.Context) {
	document := entity.Document{}
	resolveParam(ctx, &document)
	userId := middleware.CurrentUserId(ctx)
	ctx.JSON(common.NewSuccessData("查询成功", service.DocumentList(document.BookId, userId)))
}

// 查询文档
func DocumentGet(ctx iris.Context) {
	document := entity.Document{}
	resolveParam(ctx, &document)
	userId := middleware.CurrentUserId(ctx)
	ctx.JSON(common.NewSuccessData("查询成功", service.DocumentGet(document.Id, userId)))
}

// 查询公开发布文档
func DocumentGetPublished(ctx iris.Context) {
	id := ctx.Params().Get("id")
	ctx.JSON(common.NewSuccessData("查询成功", service.DocumentGetPublished(id)))
}

// 分页查询公开发布文档列表
func DocumentPagePublished(ctx iris.Context) {
	pageCondition := common.PageCondition[entity.DocumentPageCondition]{}
	resolveParam(ctx, &pageCondition)
	ctx.JSON(common.NewSuccessData("查询成功", service.DocumentPagePublished(pageCondition)))
}
