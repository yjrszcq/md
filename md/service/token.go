package service

import (
	"fmt"
	"md/dao"
	"md/middleware"
	"md/model/common"
	"md/model/entity"
	"md/util"
	"time"

	"github.com/muesli/cache2go"
)

const AccessTokenExpire = time.Hour
const RefreshTokenExpire = time.Hour * 24 * 180

// 注册
func SignUp(user entity.User) {
	tx := middleware.DbW.MustBegin()
	defer tx.Rollback()

	// 如不允许注册，查询是否没有任何用户
	if !common.Register {
		commonResult, err := dao.UserCount(tx)
		if err != nil {
			panic(common.NewErr("注册失败", err))
		}
		if commonResult.Count > 0 {
			panic(common.NewError("暂不支持注册"))
		}
	}

	// 去除用户名的空白
	user.Name = util.RemoveBlank(user.Name)
	if user.Name == "" || user.Password == "" {
		panic(common.NewError("用户名或密码不可为空"))
	}

	// 用户名长度限制
	if util.StringLength(user.Name) > 30 {
		panic(common.NewError("用户名不可大于30个字符"))
	}

	// 查询用户名不可重复
	commonResult, err := dao.UserCountByName(tx, user.Name)
	if err != nil {
		panic(common.NewErr("注册失败", err))
	}
	if commonResult.Count > 0 {
		panic(common.NewError("用户名已被注册"))
	}

	// 保存用户信息
	user.Id = util.SnowflakeString()
	user.Password = util.EncryptSHA256([]byte(user.Id + user.Password))
	user.CreateTime = time.Now().UnixMilli()
	dao.UserAdd(tx, user)

	err = tx.Commit()
	if err != nil {
		panic(common.NewErr("注册失败", err))
	}
}

// 登录
func SignIn(user entity.User) common.TokenResult {
	// 去除用户名的空白
	user.Name = util.RemoveBlank(user.Name)
	if user.Name == "" || user.Password == "" {
		panic(common.NewError("用户名或密码不可为空"))
	}

	// 校验登录次数
	checkSignInTimes(user.Name)

	// 根据用户名查询用户
	userResult, err := dao.UserGetByName(middleware.Db, user.Name)
	if err != nil {
		panic(common.NewErr("用户名或密码错误", err))
	}

	// 匹配密码：sha256(id + password)
	if util.EncryptSHA256([]byte(userResult.Id+user.Password)) != userResult.Password {
		panic(common.NewError("用户名或密码错误"))
	}

	// 生成token
	tokenResult := common.TokenResult{}
	tokenResult.Name = userResult.Name
	tokenResult.AccessToken = util.RandomString(64)
	tokenResult.RefreshToken = util.RandomString(64)

	tokenCache := common.TokenCache{}
	tokenCache.Id = userResult.Id
	tokenCache.TokenResult = tokenResult

	// 缓存token
	cache2go.Cache(common.AccessTokenCache).Add(tokenResult.AccessToken, AccessTokenExpire, &tokenCache)
	cache2go.Cache(common.RefreshTokenCache).Add(tokenResult.RefreshToken, RefreshTokenExpire, &tokenCache)
	cache2go.Cache(common.SignInTimesCache).Delete(user.Name)

	return tokenResult
}

// 退出登录
func SignOut(tokenResult common.TokenResult) {
	res, err := cache2go.Cache(common.RefreshTokenCache).Value(tokenResult.RefreshToken)
	if err == nil {
		tokenCache := res.Data().(*common.TokenCache)
		if tokenCache.RefreshToken != "" {
			cache2go.Cache(common.RefreshTokenCache).Delete(tokenCache.RefreshToken)
		}
		if tokenCache.AccessToken != "" {
			cache2go.Cache(common.AccessTokenCache).Delete(tokenCache.AccessToken)
		}
	}
}

// 刷新token
func TokenRefresh(refreshToken string) common.TokenResult {
	res, err := cache2go.Cache(common.RefreshTokenCache).Value(refreshToken)
	if err != nil {
		panic(common.NewError("认证信息已过期，请重新登录"))
	}
	tokenCache := res.Data().(*common.TokenCache)
	if tokenCache.RefreshToken == "" {
		panic(common.NewError("认证信息已过期，请重新登录"))
	}

	// Delete old tokens before generating new ones
	if tokenCache.AccessToken != "" {
		cache2go.Cache(common.AccessTokenCache).Delete(tokenCache.AccessToken)
	}
	cache2go.Cache(common.RefreshTokenCache).Delete(refreshToken)

	// 重新生成token
	tokenResult := common.TokenResult{}
	tokenResult.Name = tokenCache.Name
	tokenResult.AccessToken = util.RandomString(64)
	tokenResult.RefreshToken = util.RandomString(64)

	newTokenCache := common.TokenCache{}
	newTokenCache.Id = tokenCache.Id
	newTokenCache.TokenResult = tokenResult

	// 缓存token
	cache2go.Cache(common.AccessTokenCache).Add(newTokenCache.AccessToken, AccessTokenExpire, &newTokenCache)
	cache2go.Cache(common.RefreshTokenCache).Add(newTokenCache.RefreshToken, RefreshTokenExpire, &newTokenCache)

	return tokenResult
}

// 校验登录次数，如已超出则抛出异常
func checkSignInTimes(name string) {
	cache := cache2go.Cache(common.SignInTimesCache)
	signInTimes, err := cache.Value(name)
	times := 1
	expireSecond := int64(0)
	if err != nil {
		cache.Add(name, time.Minute*5, true)
	} else {
		expireSecond = 300 - (time.Now().Unix() - signInTimes.CreatedOn().Unix())
		// 有时过期后缓存不会立即删除，手动重置次数
		if expireSecond <= 0 {
			cache.Add(name, time.Minute*5, true)
		} else {
			times = int(signInTimes.AccessCount()) + 1
		}
	}
	if times > 5 {
		panic(common.NewError(fmt.Sprintf("登录次数已达上限，请于%v分钟后再试", (expireSecond/60 + 1))))
	}
}
