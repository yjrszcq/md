package controller

import (
	"md/middleware"
	"md/model/common"

	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris/v12"
)

// 初始化iris路由
func InitRouter(app *iris.Application) {
	// 允许跨域
	app.UseRouter(cors.AllowAll())

	app.PartyFunc("/api", func(api iris.Party) {
		// 开放接口
		api.PartyFunc("/open", func(open iris.Party) {
			open.Get("/doc/get/{id}", DocumentGetPublished)
			open.Post("/doc/page", DocumentPagePulished)
		})

		// token相关接口
		api.PartyFunc("/token", func(token iris.Party) {
			token.Use(middleware.TokenAuth)

			token.Post("/sign-up", SignUp)
			token.Post("/sign-in", SignIn)
			token.Post("/sign-out", SignOut)
			token.Post("/refresh", TokenRefresh)
		})

		// 数据接口
		api.PartyFunc("/data", func(data iris.Party) {
			data.Use(middleware.DataAuth)

			data.PartyFunc("/user", func(user iris.Party) {
				user.Post("/update-password", UserUpdatePassword)
			})

			data.PartyFunc("/book", func(book iris.Party) {
				book.Post("/add", BookAdd)
				book.Post("/update", BookUpdate)
				book.Post("/delete", BookDelete)
				book.Post("/list", BookList)
			})

			data.PartyFunc("/doc", func(doc iris.Party) {
				doc.Post("/add", DocumentAdd)
				doc.Post("/update", DocumentUpdate)
				doc.Post("/update-content", DocumentUpdateContent)
				doc.Post("/delete", DocumentDelete)
				doc.Post("/list", DocumentList)
				doc.Post("/get", DocumentGet)
			})

			data.PartyFunc("/pic", func(pic iris.Party) {
				pic.Post("/page", PicturePage)
				pic.Post("/delete", PictureDelete)
				pic.Post("/upload", PictureUpload)
			})

			data.PartyFunc("/rsa", func(rsa iris.Party) {
				rsa.Post("/generate", RSAGenerateKey)
				rsa.Post("/encrypt", RSAEncrypt)
				rsa.Post("/decrypt", RSADecrypt)
				rsa.Post("/sign", RSASign)
				rsa.Post("/verify", RSAVerify)
			})

			data.PartyFunc("/ai", func(ai iris.Party) {
				ai.Get("/config", AIConfigGet)
				ai.Post("/config", AIConfigSave)
				ai.Post("/config/delete", AIConfigDelete)
				ai.Get("/config/exists", AIConfigExists)
				ai.Get("/config/full", AIConfigGetFull)
			})
		})
	})
}

// 解析参数
func resolveParam(ctx iris.Context, con interface{}) {
	err := ctx.ReadJSON(&con)
	if err != nil {
		panic(common.NewErr("参数解析失败", err))
	}
}
