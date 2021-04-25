package rtmprelay

import (
	"fmt"
	"sync"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/configure"
	"github.com/gwuhaolin/livego/protocol/rtmp/core"

	log "github.com/sirupsen/logrus"
)

type StaticPush struct {
	RtmpUrl       string
	packet_chan   chan *av.Packet
	sndctrl_chan  chan string
	connectClient *core.ConnClient
	startflag     bool
}

var G_StaticPushMap = make(map[string](*StaticPush))
var g_MapLock = new(sync.RWMutex)
var G_PushUrlList []string = nil

var (
	STATIC_RELAY_STOP_CTRL = "STATIC_RTMPRELAY_STOP"
)

// GetStaticPushList 获取静态推送的列表
func GetStaticPushList(appname string) ([]string, error) {
	if G_PushUrlList == nil {
		// Do not unmarshel the config every time, lots of reflect works -gs
		pushurlList, ok := configure.GetStaticPushUrlList(appname)
		if !ok {
			G_PushUrlList = []string{}
		} else {
			G_PushUrlList = pushurlList
		}
	}

	if len(G_PushUrlList) == 0 {
		return nil, fmt.Errorf("no static push url")
	}

	return G_PushUrlList, nil
}

// GetAndCreateStaticPushObject 创建并返回静态推送对象
func GetAndCreateStaticPushObject(rtmpurl string) *StaticPush {
	g_MapLock.RLock()
	staticpush, ok := G_StaticPushMap[rtmpurl]
	log.Debugf("GetAndCreateStaticPushObject: %s, return %v", rtmpurl, ok)
	if !ok {
		g_MapLock.RUnlock()
		newStaticpush := NewStaticPush(rtmpurl)

		g_MapLock.Lock()
		G_StaticPushMap[rtmpurl] = newStaticpush
		g_MapLock.Unlock()

		return newStaticpush
	}
	g_MapLock.RUnlock()

	return staticpush
}

// GetStaticPushObject 获取静态推送对象，没有返回错误
func GetStaticPushObject(rtmpurl string) (*StaticPush, error) {
	g_MapLock.RLock()
	if staticpush, ok := G_StaticPushMap[rtmpurl]; ok {
		g_MapLock.RUnlock()
		return staticpush, nil
	}
	g_MapLock.RUnlock()

	return nil, fmt.Errorf("G_StaticPushMap[%s] not exist....", rtmpurl)
}

// ReleaseStaticPushObject 正式的静态推送对象
func ReleaseStaticPushObject(rtmpurl string) {
	g_MapLock.RLock()
	if _, ok := G_StaticPushMap[rtmpurl]; ok {
		g_MapLock.RUnlock()

		log.Debugf("ReleaseStaticPushObject %s ok", rtmpurl)
		g_MapLock.Lock()
		delete(G_StaticPushMap, rtmpurl)
		g_MapLock.Unlock()
	} else {
		g_MapLock.RUnlock()
		log.Debugf("ReleaseStaticPushObject: not find %s", rtmpurl)
	}
}

// NewStaticPush 创建新的静态推送
func NewStaticPush(rtmpurl string) *StaticPush {
	return &StaticPush{
		RtmpUrl:       rtmpurl,
		packet_chan:   make(chan *av.Packet, 500),
		sndctrl_chan:  make(chan string),
		connectClient: nil,
		startflag:     false,
	}
}

// Start 开始静态推送
func (self *StaticPush) Start() error {
	if self.startflag {
		return fmt.Errorf("StaticPush already start %s", self.RtmpUrl)
	}

	self.connectClient = core.NewConnClient()

	log.Debugf("static publish server addr:%v starting....", self.RtmpUrl)
	err := self.connectClient.Start(self.RtmpUrl, "publish")
	if err != nil {
		log.Debugf("connectClient.Start url=%v error", self.RtmpUrl)
		return err
	}
	log.Debugf("static publish server addr:%v started, streamid=%d", self.RtmpUrl, self.connectClient.GetStreamId())
	go self.HandleAvPacket()

	self.startflag = true
	return nil
}

// Stop 停止推送
func (self *StaticPush) Stop() {
	if !self.startflag {
		return
	}

	log.Debugf("StaticPush Stop: %s", self.RtmpUrl)
	self.sndctrl_chan <- STATIC_RELAY_STOP_CTRL
	self.startflag = false
}

// WriteAvPacket 写av包
func (self *StaticPush) WriteAvPacket(packet *av.Packet) {
	if !self.startflag {
		return
	}

	self.packet_chan <- packet
}

// sendPacket 发送包
func (self *StaticPush) sendPacket(p *av.Packet) {
	if !self.startflag {
		return
	}
	var cs core.ChunkStream

	cs.Data = p.Data
	cs.Length = uint32(len(p.Data))
	cs.StreamID = self.connectClient.GetStreamId()
	cs.Timestamp = p.TimeStamp
	//cs.Timestamp += v.BaseTimeStamp()

	//log.Printf("Static sendPacket: rtmpurl=%s, length=%d, streamid=%d",
	//	self.RtmpUrl, len(p.Data), cs.StreamID)
	if p.IsVideo {
		cs.TypeID = av.TAG_VIDEO
	} else {
		if p.IsMetadata {
			cs.TypeID = av.TAG_SCRIPTDATAAMF0
		} else {
			cs.TypeID = av.TAG_AUDIO
		}
	}

	self.connectClient.Write(cs)
}

// HandleAvPacket 处理包
func (self *StaticPush) HandleAvPacket() {
	if !self.IsStart() {
		log.Debugf("static push %s not started", self.RtmpUrl)
		return
	}

	for {
		select {
		case packet := <-self.packet_chan:
			self.sendPacket(packet)
		case ctrlcmd := <-self.sndctrl_chan:
			if ctrlcmd == STATIC_RELAY_STOP_CTRL {
				self.connectClient.Close(nil)
				log.Debugf("Static HandleAvPacket close: publishurl=%s", self.RtmpUrl)
				return
			}
		}
	}
}

// IsStart 是否已启动
func (self *StaticPush) IsStart() bool {
	return self.startflag
}
