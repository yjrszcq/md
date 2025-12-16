package middleware

import (
	"os"
)

// 初始化数据目录
func InitDataDir(dataPath, resourceName, pictureName, thumbnailName string) error {
	path := dataPath + resourceName

	err := os.MkdirAll(path+"/"+pictureName, 0755)
	if err != nil {
		Log.Error("创建图片目录失败：", err)
		return err
	}

	err = os.MkdirAll(path+"/"+thumbnailName, 0755)
	if err != nil {
		Log.Error("创建缩略图目录失败：", err)
		return err
	}

	return nil
}
