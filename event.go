package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/aluka-7/configuration/backends"
)

// Event 系统之间的准实时通知事件对象，封装需要发布的事件信息。
// 需要特别注意的是，每个事件的key必须是全局唯一的，不仅仅是系统内唯一，而是整个系统生态内唯一。
// 另外，每个时间会被赋予一个超时时间（默认是10s），当事件被发布时会被设置一个发布时间，接收端接收到事件后会根据其系统本地时间和发布时间进行对
// 比，如果差额超过了超时时间则认为事件超时了则事件监听程序不会被执行，反之事件程序会得到执行。
// 每个事件允许携带一定量的数据，携带的数据总量以字节长度计算(JSON格式化后的)，限制最大数据长度为1k，超过则不允许发送事件。
type Event struct {
	Key       string                 `json:"key"`       // 事件的唯一key
	PubTime   int64                  `json:"pub_time"`  // 事件的发布时间，单位毫秒
	Timeout   int64                  `json:"timeout"`   // 超时时间，单位毫秒
	Body      map[string]interface{} `json:"body"`      // 事件的数据
	Published bool                   `json:"published"` // 是否发布了，防止重复发送
}

// AddData 添加一个key(数据对应的key)-value(数据对应的值)数据到事件对象中。
func (e *Event) AddData(key string, value interface{}) *Event {
	e.Body[key] = value
	return e
}

// SetPubTime 设置事件的pubTime(发布时间)，当事件被发布时由系统自动设置。
func (e *Event) SetPubTime(pubTime int64) *Event {
	e.PubTime = pubTime
	return e
}

// SetPublished 标示该事件对象已被发布过。
func (e *Event) SetPublished() *Event {
	e.SetPubTime(time.Now().UnixNano())
	e.Published = true
	return e
}

// GeKey 获取当前事件对象对应的事件key，总是不为空。
func (e Event) GeKey() string {
	return e.Key
}

// GetData 获取指定key的事件数据，如果指定的数据key不存在则会返回nil。
func (e *Event) GetData(key string) interface{} {
	v, ok := e.Body[key]
	if ok {
		return v
	} else {
		return nil
	}
}

// GePubTime 获取时间的发布时间，单位毫秒。
func (e Event) GePubTime() int64 {
	return e.PubTime
}

// IsTimeout 判断当前事件是否超时了，以系统本地时间和发布时间的差额为判断标准。
func (e Event) IsTimeout() bool {
	return time.Now().UnixNano()-e.PubTime > e.Timeout
}

// IsPublished 事件是否被发布了，防止重复发送。
func (e Event) IsPublished() bool {
	return e.Published
}
func (e Event) Json() ([]byte, error) {
	return json.Marshal(e)
}

// IsOverloaded 判断当前事件数据载体的总字节大小是否超标了，以JSON格式化后的字节数为标准，最大只允许1024字节的数据。
func (e *Event) IsOverloaded() bool {
	if b, err := json.Marshal(e); err != nil {
		return false
	} else {
		return bytes.Count(b, nil)-1 > 1024
	}
}

// NewFCEventTimeout 给定事件的key和超时时间来构建一个事件对象，每个事件都有一个发布时间，如果接收时间的系统时间-发布时间>超时时间则认为事件超时了，则接收端系统不会被触发事件处理程序。
func NewFCEventTimeout(key string, timeout int64) (*Event, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("事件key不能为空")
	} else {
		reg, _ := regexp.Compile("[a-z]+([a-z0-9-])*")
		s := reg.FindStringSubmatch(key)
		if len(s) == 0 {
			return nil, fmt.Errorf("事件key不符合规范")
		} else {
			return &Event{Key: key, Timeout: timeout, Body: make(map[string]interface{}, 0)}, nil
		}
	}
}

// NewFCEvent 使用事件key构造一个具有默认超时时间的事件对象，默认的超时时间为10s。
func NewFCEvent(key string) (*Event, error) {
	return NewFCEventTimeout(key, int64(10000))
}

// EventListener 跨系统的事件监听器接口定义，适用于两个/多个在线系统之间的实时通知，业务系统只需要实现该接口并注册到服务中后即可实现跨系统的事件监听。
type EventListener interface {
	// EventKeys 该监听器要监听的事件key列表(可指定多个)，总是返回非空列表。
	EventKeys() []string

	// OnEvent 事件监听到的回调处理方法，由业务系统自行处理。
	OnEvent(event Event)
}

func EventEngine(conf backends.StoreConfig) eventEngine {
	fmt.Println("Loading Aluka Event Engine")
	store, err := backends.New(conf)
	if err != nil {
		panic(err)
	}
	return eventEngine{store, "/system_events"}
}

type eventEngine struct {
	store     backends.StoreClient
	eventPath string
}

func (ee eventEngine) Publish(e *Event) error {
	if e == nil {
		return fmt.Errorf("发布的事件不能为nil")
	}
	if e.IsPublished() {
		return fmt.Errorf("事件已被发布过")
	}
	if e.IsOverloaded() {
		return fmt.Errorf("事件的载体数据超标")
	}
	e.SetPublished()
	// 如果不存在该路径则先创建，然后再设置数据
	path := ee.eventPath + "/" + e.Key
	if b, e := e.Json(); e == nil {
		_, er := ee.store.Add(path, b)
		return er
	} else {
		return e
	}
}
func (ee eventEngine) StartEventListener(listener []EventListener) {
	p := &parser{}
	p.Init(listener)
	go WatchProcessor([]string{"/system_events"}, ee.store).Process(p)
}

type parser struct {
	listenerMap map[string][]EventListener
}

func (p parser) Init(listener []EventListener) {
	p.listenerMap = make(map[string][]EventListener, 0)
	for _, v := range listener {
		if len(v.EventKeys()) == 0 {
			continue
		}
		for _, e := range v.EventKeys() {
			list := p.listenerMap[e]
			list = append(list, v)
			p.listenerMap[e] = list
		}
	}
}
func (p parser) Changed(data map[string]string) {
	fmt.Printf("data:%+v\n", data)
	for k, v := range data {
		for _, val := range p.listenerMap[k] {
			var e Event
			if err := json.Unmarshal([]byte(v), &e); err == nil {
				val.OnEvent(e)
			}
		}
	}
}
