package backends

import (
	"github.com/aluka-7/configuration/backends/mock"
	"github.com/aluka-7/configuration/backends/zookeeper"
	"github.com/samuel/go-zookeeper/zk"
)

type StoreConfig struct {
	Backend      string            `json:"backend"`      // 后端服务地址
	Username     string            `json:"username"`     // 只读用户名
	Password     string            `json:"password"`     // 只读用户密码
	OpenUser     string            `json:"openUser"`     // 可读写用户名
	OpenPassword string            `json:"openPassword"` // 可读写密码
	Exp          map[string]string `json:"exp"`
}

// The StoreClient interface is implemented by objects that can retrieve key/value pairs from a backend store.
type StoreClient interface {
	Client() *zk.Conn
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
	Lock(path string) *zk.Lock
	Add(path string, value []byte, flags int32) (string, error)
	Modify(path string, value []byte) error
	Delete(path string) error
}

// New is used to create a storage client based on our configuration.
func New(conf StoreConfig) (StoreClient, error) {
	return zookeeper.NewZookeeperClient([]string{conf.Backend}, conf.Username, conf.Password, conf.OpenUser, conf.OpenPassword)
}
func NewMock(conf StoreConfig) (StoreClient, error) {
	return mock.NewMockClient(conf.Exp)
}
