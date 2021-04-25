// Package mp4
// 
// @author: xwc1125
// @date: 2021/4/25
package mp4

import (
	log "github.com/sirupsen/logrus"
	"github.com/xfrr/goffmpeg/transcoder"
	"os"
)

type fileInfo struct {
	inputPath  string
	outputPath string
}

type Mp4 struct {
	transcoder *transcoder.Transcoder
	convertCh  chan *fileInfo
}

func New() *Mp4 {
	trans := new(transcoder.Transcoder)
	m := &Mp4{
		transcoder: trans,
		convertCh:  make(chan *fileInfo, 100),
	}
	go m.listen()
	return m
}

func (m *Mp4) Add(inputPath string, outputPath string) {
	m.convertCh <- &fileInfo{
		inputPath:  inputPath,
		outputPath: outputPath,
	}
}
func (m *Mp4) listen() {
	for {
		select {
		case ch := <-m.convertCh:
			m.Convert(ch.inputPath, ch.outputPath)
		}
	}
}

func (m *Mp4) Convert(inputPath string, outputPath string) error {
	log.Debugf("start mp4 convert.inputPath=%s,outputPath=%s", inputPath, outputPath)
	outputPathTemp := inputPath + "_temp.mp4"
	err := m.transcoder.Initialize(inputPath, outputPathTemp)
	if err != nil {
		return err
	}
	done := m.transcoder.Run(false)
	err = <-done
	if err != nil {
		log.Error(err)
	}
	os.Rename(outputPathTemp, outputPath)
	log.Debugf("end mp4 convert.inputPath=%s,outputPath=%s", inputPath, outputPath)
	return err
}
