package mock

import (
	"github.com/samuel/go-zookeeper/zk"
)

func NewMockClient(store map[string]string) (*Client, error) {
	return &Client{store: store}, nil
}

type Client struct {
	store map[string]string
}

func (c Client) Client() *zk.Conn {
	return nil
}

func (c Client) Lock(path string) *zk.Lock {
	return nil
}

func (c Client) Delete(path string) error {
	return nil
}

func (c Client) Add(path string, value []byte, flags int32) (string, error) {
	return "", nil
}

func (c Client) Modify(path string, value []byte) error {
	return nil
}

func (c Client) GetValues(keys []string) (vls map[string]string, err error) {
	vls = make(map[string]string, len(keys))
	for _, v := range keys {
		vls[v] = c.store[v]
	}
	return
}

func (c Client) WatchPrefix(keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	return 0, nil
}
