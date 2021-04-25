// Package m3u8tomp4
// 
// @author: xwc1125
// @date: 2021/3/11
package tsmerge

import "testing"

func TestTsMerge(t *testing.T) {
	//err := TsMerge("http://127.0.0.1:8081/hls/L17LTlsVqMNTZyLKMIFSD2x28MlgPJ0SDZVHnHJPxMKi0t12/record.m3u8", "name1")
	err := TsMerge("/Users/yijiaren/Workspaces/git/github/go/livego-center/lal/tmp/lal/hls/L17LTlsVqMNTZyLKMIFSD2x28MlgPJ0SDZVHnHJPxMKi0t12/record.m3u8", "./test1/name2")
	if err != nil {
		panic(err)
	}
}
