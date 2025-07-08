package utils

import (
  "archive/zip"
  "bytes"
  "github.com/cloudwego/hertz/pkg/common/hlog"
  "golang.org/x/text/encoding/simplifiedchinese"
  "golang.org/x/text/transform"
  "io"
  "io/ioutil"
  "mime/multipart"
  "os"
  "path/filepath"
)

// 移动文件
func MoveFile(file multipart.File, movedDir string, dstFilename string) (bool, error) {
  // 创建移动路径
  err := os.MkdirAll(movedDir, 0755)
  if err != nil {
    return true, err
  }
  //移动到这个文件夹
  dstFilePath := movedDir + "/" + dstFilename
  hlog.Info("dstFilePath:", dstFilePath)
  err = os.Remove(dstFilePath)
  if err != nil {
    hlog.Error(err.Error())
    if os.IsNotExist(err) {
      hlog.Info("ignore this error")
    } else {
      return true, err
    }

  }
  //创建文件,移动文件只需要创建,如果添加其他选项会出现错误
  openFile, err := os.OpenFile(dstFilePath, os.O_CREATE|os.O_WRONLY, 0666)
  if err != nil {
    hlog.Error(err.Error())
    return true, err
  }
  defer openFile.Close()

  n, err := io.Copy(openFile, file)
  if err != nil {
    hlog.Error(err.Error())
    return true, err
  } else {
    hlog.Info(n)
  }
  return false, nil
}

// 检查文件或目录是否存在
func fileExists(path string) bool {
  _, err := os.Stat(path)
  if os.IsNotExist(err) {
    return false
  }
  return err == nil
}

// 如果文件存在则删除
func removeFileIfExists(path string) error {
  if fileExists(path) {
    err := os.RemoveAll(path)
    if err != nil {
      return err
    }
  }
  return nil
}
func ExtractFile(file multipart.File, targetDir string, length int64) (bool, error) {
  // 创建解压路径
  err := removeFileIfExists(targetDir)
  if err != nil {
    hlog.Error("err:", err)
    return true, err
  }
  err = os.MkdirAll(targetDir, 0755)
  if err != nil {
    hlog.Error("err:", err)
    return true, err
  }

  // 解压文件
  reader, err := zip.NewReader(file, length)
  if err != nil {
    hlog.Error("err:", err)
    return true, err
  }

  for _, f := range reader.File {
    filename := GetChineseName(f.Name)
    hlog.Info("filename:", filename)
    path := filepath.Join(targetDir, filename)

    var err error = nil
    if f.FileInfo().IsDir() {
      err = os.MkdirAll(path, f.Mode())
      if err != nil {
        hlog.Error("err:", err)
        return true, err
      }
      continue
    }

    err = os.MkdirAll(filepath.Dir(path), 0755)
    if err != nil {
      hlog.Error("err:", err)
      return true, err
    }

    unzippedFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
    if err != nil {
      hlog.Error("err:", err)
      return true, err
    }
    defer unzippedFile.Close()

    zippedFile, err := f.Open()
    if err != nil {
      hlog.Error("err:", err)
      return true, err
    }
    defer zippedFile.Close()

    _, err = io.Copy(unzippedFile, zippedFile)
    if err != nil {
      hlog.Error("err:", err)
      return true, err
    }
  }
  return false, nil
}

// 获取中文名称
func GetChineseName(filename string) string {
  i := bytes.NewReader([]byte(filename))
  decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
  content, _ := ioutil.ReadAll(decoder)
  filename = string(content)
  return filename
}
