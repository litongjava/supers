package services

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/litongjava/supers/model"
	"github.com/litongjava/supers/utils"
	"os"
	"path/filepath"
)

func RunWrapperCommand(workDir string, command string) (model.CommandResult, error) {
	if "nginx-reload" == command {
		return utils.RunComamnd(workDir, "nginx", "-s", "reload")
	} else if "nginx-t" == command {
		return utils.RunComamnd(workDir, "nginx", "-t")
	} else {
		//创建一个文件
		hash := md5.New()
		hash.Write([]byte(command))
		md5String := hex.EncodeToString(hash.Sum(nil))
		movedDir := filepath.Join(workDir, "script")
		// 创建移动路径
		err := os.MkdirAll(movedDir, 0755)
		if err != nil {
			result := model.CommandResult{}
			result.Success = false
			return result, err
		}
		//移动到这个文件夹
		dstFilePath := filepath.Join(movedDir, md5String)
		hlog.Info("dstFilePath:", dstFilePath)
		//创建文件
		scriptFile, err := os.OpenFile(dstFilePath, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			result := model.CommandResult{}
			result.Success = false
			return result, err
		}
		defer scriptFile.Close()
		_, err = scriptFile.WriteString(command)
		if err != nil {
			result := model.CommandResult{}
			result.Success = false
			return result, err
		}
		return utils.RunComamnd(workDir, "sh", dstFilePath)
	}
}
