package zookeeper

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/samuel/go-zookeeper/zk"
)

const DefaultSessionTimeout = time.Second * 10

// Client provides a wrapper around the zookeeper client
type Client struct {
	client *zk.Conn
}

type watchResponse struct {
	waitIndex uint64
	err       error
}

func NewZookeeperClient(machines []string, user, password, openUser, openPassword string) (*Client, error) {
	c, _, err := zk.Connect(machines, DefaultSessionTimeout)
	if err != nil {
		panic(err)
	}
	if err := c.AddAuth("digest", []byte(user+":"+password)); err != nil {
		log.Err(err).Msg("AddAuth user returned error")
	}
	if len(openUser) > 0 {
		if err := c.AddAuth("digest", []byte(openUser+":"+openPassword)); err != nil {
			log.Err(err).Msg("AddAuth openUser returned error")
		}
	}
	return &Client{c}, nil
}

func (c *Client) Client() *zk.Conn {
	return c.client
}

func (c *Client) Lock(path string) *zk.Lock {
	return zk.NewLock(c.client, path, zk.WorldACL(zk.PermAll))
}

func (c *Client) Add(path string, value []byte, flags int32) (string, error) {
	// flags有4种取值：
	// 0:永久，除非手动删除
	// zk.FlagEphemeral = 1:短暂，session断开则该节点也被删除
	// zk.FlagSequence  = 2:会自动在节点后面添加序号
	// 3:Ephemeral和Sequence，即，短暂且自动添加序号
	return c.client.Create(path, value, flags, zk.WorldACL(zk.PermAll))
}

func (c *Client) Modify(path string, value []byte) error {
	_, stat, _ := c.client.Get(path)
	_, err := c.client.Set(path, value, stat.Version)
	return err
}

func (c *Client) Delete(path string) error {
	_, stat, _ := c.client.Get(path)
	err := c.client.Delete(path, stat.Version)
	return err
}

func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, v := range keys {
		_, _, err := c.client.Exists(v)
		if err != nil {
			return vars, err
		}
		if b, _, err := c.client.Get(v); err != nil {
			return vars, err
		} else {
			vars[v] = string(b)
		}
	}
	return vars, nil
}

func (c *Client) WatchPrefix(keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	// return something > 0 to trigger a key retrieval from the store
	if waitIndex == 0 {
		return 1, nil
	}

	respChan := make(chan watchResponse)
	cancelRoutine := make(chan bool)
	defer close(cancelRoutine)

	// watch all keys in prefix for changes
	for _, v := range keys {
		log.Info().Msgf("Watching:%s", v)
		go c.watch(v, respChan, cancelRoutine)
		break
	}

	for {
		select {
		case <-stopChan:
			return waitIndex, nil
		case r := <-respChan:
			return r.waitIndex, r.err
		}
	}
}

func (c *Client) watch(key string, respChan chan watchResponse, cancelRoutine chan bool) {
	_, _, keyEventCh, err := c.client.GetW(key)
	if err != nil {
		respChan <- watchResponse{0, err}
	}
	_, _, childEventCh, err := c.client.ChildrenW(key)
	if err != nil {
		respChan <- watchResponse{0, err}
	}
	select {
	case <-cancelRoutine:
		log.Info().Msgf("Stop watching:%s", key)
		return
	case e := <-keyEventCh:
		if e.Type == zk.EventNodeDataChanged {
			respChan <- watchResponse{1, e.Err}
		} else if e.Type == zk.EventNotWatching {
			log.Info().Msgf("Not watching:%s", key)
			respChan <- watchResponse{0, e.Err}
		}
	case e := <-childEventCh:
		if e.Type == zk.EventNodeChildrenChanged {
			respChan <- watchResponse{1, e.Err}
		} else if e.Type == zk.EventNotWatching {
			log.Info().Msgf("Not watching:%s", key)
			respChan <- watchResponse{0, e.Err}
		}
	}
}
