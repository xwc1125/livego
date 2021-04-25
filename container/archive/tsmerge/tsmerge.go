// Package m3u8tomp4
// 
// @author: xwc1125
// @date: 2021/3/11
package tsmerge

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// TsMerge 将m3u8转换为mp4
// m3u8Url：可以是url，文件路径
// filename：自定义的文件名称，不需要含有.ts
func TsMerge(m3u8Url string, mp4Name string) error {
	// 判断文件的结尾是否为.m3u8
	if !strings.HasSuffix(m3u8Url, ".m3u8") {
		fmt.Println("please input correct m3u8 file url!!")
		return errors.New("please input correct m3u8 file url!!")
	}

	var (
		ts  = []string{}
		err error
	)
	filePath := m3u8Url
	// 获取文件名[download是拿到m3u8文件的内容]
	if isHttp(m3u8Url) {
		filePath, err = downloadFile(m3u8Url)
		if err != nil {
			return err
		}
	}
	fmt.Println("filename", filePath)
	// 解析m3u8文件,获取ts的所有文件
	ts = parseM3u8(filePath)

	// 创建mp4
	if mp4Name == "" {
		mp4Name = filePath
	}
	if !strings.HasPrefix(mp4Name, ".ts") {
		mp4Name = mp4Name + ".ts"
		basePath := filepath.Dir(mp4Name)
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			os.MkdirAll(basePath, 0777)
		}
		os.Create(mp4Name)
	}
	fd, err := os.OpenFile(mp4Name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()

	var prefix string
	if isHttp(m3u8Url) {
		splitAfter := strings.Split(m3u8Url, "/")
		prefix = strings.Replace(m3u8Url, splitAfter[len(splitAfter)-1], "", 1)
	} else {
		prefix, _ = filepath.Split(m3u8Url)
	}

	for _, tsUrl := range ts {
		if isHttp(m3u8Url) {
			// 链接是网址
			if isHttp(tsUrl) {
				if err = writeHttpData(fd, tsUrl); err != nil {
					return err
				}
			} else {
				if err = writeHttpData(fd, prefix+tsUrl); err != nil {
					return err
				}
			}
		} else {
			// 读本地数据
			if filepath.IsAbs(tsUrl) {
				if err = writeFileData(fd, tsUrl); err != nil {
					return err
				}
			} else {
				if err = writeFileData(fd, path.Join(prefix, tsUrl)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func isHttp(url string) bool {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return true
	}
	return false
}

// 写网络数据
func writeHttpData(fd *os.File, url string) error {
	rs, err := http.Get(url)
	if err != nil {
		return err
	}
	io.Copy(fd, rs.Body)
	return nil
}

// writeFileData 写文件
func writeFileData(fd *os.File, tsFile string) error {
	f, err := os.OpenFile(tsFile, os.O_RDONLY, 0600)
	defer f.Close()
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	_, err = fd.Write(bytes)
	return err
}

//parseM3u8 解析m3u8文件[拿到ts的所有文件]
func parseM3u8(m3u8Path string) []string {
	var ret = []string{}

	handler, err := os.Open(m3u8Path)
	if err != nil {
		fmt.Println("Parse m3u8 file error ", err)
	}
	defer handler.Close()

	buffer := bufio.NewReader(handler)
	for {
		line, _, err := buffer.ReadLine()
		if err == io.EOF {
			break
		}
		s := string(line)
		//if strings.HasPrefix(s, "http") {
		//	ret = append(ret, s)
		//}
		if strings.HasSuffix(s, ".ts") {
			ret = append(ret, s)
		}
	}

	return ret
}

// downloadFile 下载ts文件
func downloadFile(path string) (string, error) {
	var pathinfo = strings.Split(path, "/")
	var filename = pathinfo[len(pathinfo)-1]
	res, err := http.Get(path)
	if err != nil {
		return "", err
	}

	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}

	io.Copy(f, res.Body)
	return filename, nil
}
