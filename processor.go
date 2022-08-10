package configuration

import (
	"sync"
	"time"

	"github.com/aluka-7/configuration/backends"
)

//ChangedListener 配置数据发生变化时的监听器接口。
type ChangedListener interface {
	// Changed 指定配置变化后的通知接口，标示变化的路径和变化后的data(变化后的新数据值)。
	Changed(data map[string]string)
}

type Processor interface {
	Process(listener ChangedListener)
}
type watchProcessor struct {
	path     []string
	stopChan chan bool
	doneChan chan bool
	errChan  chan error
	wg       sync.WaitGroup
	store    backends.StoreClient
}

func WatchProcessor(path []string, store backends.StoreClient) Processor {
	stopChan := make(chan bool)
	doneChan := make(chan bool)
	errChan := make(chan error, 10)
	var wg sync.WaitGroup
	return &watchProcessor{path, stopChan, doneChan, errChan, wg, store}
}

func (p watchProcessor) Process(listener ChangedListener) {
	var lastIndex uint64
	defer close(p.doneChan)
	go p.monitorPrefix(p.path, lastIndex, listener)
	p.wg.Wait()
}

func (p watchProcessor) monitorPrefix(path []string, lastIndex uint64, listener ChangedListener) {
	defer p.wg.Done()
	for {
		index, err := p.store.WatchPrefix(path, lastIndex, p.stopChan)
		if err != nil {
			p.errChan <- err
			//防止后端错误占用所有资源.
			time.Sleep(time.Second * 2)
			continue
		}
		if lastIndex > 0 {
			if vl, err := p.store.GetValues(path); err == nil {
				listener.Changed(vl)
			} else {
				p.errChan <- err
			}
		}
		lastIndex = index
	}
}
