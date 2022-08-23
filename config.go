package configuration

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/aluka-7/configuration/backends"
	"github.com/aluka-7/utils"
	"github.com/rs/zerolog/log"
)

const Namespace = "/system"
const DesKey = "aluka-7!"

// NewStoreConfig 提供给所有业务系统使用的配置管理引擎，所有业务系统/中间件的可变配置通过集中配置中心进行统一
// 配置，业务系统可通过该类来管理 自己的配置数据并在配置中心的数据发生变化时得到及时的通知。
//
// 配置管理中心的数据存储为树形目录结构（类似文件系统的目录结构），每一个节点都可以存储相应的数据。业务系统可在基础目录树的基础上
// 往下继续扩展新的目录结构，从而可以区分存储不同的配置项信息。比如系统分配给业务系统A的基础目录（称为path）为/sysa，
// 那么业务系统可以扩充到如下path：/sysa/key1、/sysa/key2、/sysa/key3，每个path可存储对应的数据。
//
// 配置管理引擎客户端依赖连接的配置管理中心，需要在同一配置管理中心下载对应的授权文件方可。
func NewStoreConfig() (conf backends.StoreConfig) {
	uaf := os.Getenv("UAF")
	if len(uaf) == 0 {
		if b, err := ioutil.ReadFile("./configuration.uaf"); err == nil {
			uaf = string(b)
		}
	}
	if len(uaf) > 0 {
		ds, _ := base64.URLEncoding.DecodeString(uaf)
		data, er := utils.Decrypt(ds, []byte(DesKey))
		if er != nil {
			panic("解密configuration.uaf出错")
		}
		if er = json.Unmarshal(data, &conf); er != nil {
			panic("不存在配置资源请使用本地配置信息")
		}
	} else {
		panic("请在环境变量中UAF或者./configuration.uaf中配置内容")
	}
	return
}
func DefaultEngine() Configuration {
	return Engine(NewStoreConfig())
}
func MockEngine(t *testing.T, conf backends.StoreConfig) Configuration {
	fmt.Println("Loading Aluka configuration Mock Engine")
	store, err := backends.NewMock(conf)
	if err != nil {
		panic(err)
	}
	return &configuration{store}
}

//Engine 获取配置管理引擎的唯一实例。
func Engine(conf backends.StoreConfig) Configuration {
	fmt.Println("Loading Aluka configuration Engine")
	store, err := backends.New(conf)
	if err != nil {
		panic(err)
	}
	return &configuration{store}
}

type Configuration interface {
	Values(app, group, tag string, path []string) (map[string]string, error)
	String(app, group, tag, path string) (string, error)
	Clazz(app, group, tag, path string, clazz interface{}) error
	Get(app, group, tag string, path []string, parser ChangedListener)
}

type configuration struct {
	store backends.StoreClient
}

func (configuration) maskPath(app, group, tag, path string) string {
	key := []string{Namespace, app, group, path}
	if len(tag) > 0 {
		key = []string{Namespace, app, group, tag, path}
	}
	return strings.Join(key, "/")
}

// Values 获取多个配置项的配置信息，返回原始的配置数据格式(map集合)，如果获取失败则抛出异常。
// key 配置项的路径，如：/abc/dd
func (c configuration) Values(app, group, tag string, path []string) (map[string]string, error) {
	_path := make([]string, len(path))
	for i, v := range path {
		_path[i] = c.maskPath(app, group, tag, v)
	}
	vl, err := c.store.GetValues(_path)
	if err != nil {
		log.Err(err).Msgf("获取多个配置项[%v]的配置信息出错:%+v", path, err)
	} else {
		log.Info().Msgf("获取多个配置项为:%+v", vl)
	}
	return vl, err
}

//String 获取指定配置项的配置信息，返回原始的配置数据格式，如果获取失败则抛出异常。
func (c configuration) String(app, group, tag, path string) (string, error) {
	path = c.maskPath(app, group, tag, path)
	vl, err := c.store.GetValues([]string{path})
	if err != nil {
		log.Err(err).Msgf("获取多个配置项[%s]的配置信息出错:%+v", path, err)
	} else {
		log.Info().Msgf("获取多个配置项为:%+v", vl)
	}
	return vl[path], err
}

//Clazz 获取指定配置项的配置信息，并且将配置信息（JSON格式的）转换为指定的Go结构体，如果获取失败或转换失败则抛出异常。
func (c configuration) Clazz(app, group, tag, path string, clazz interface{}) error {
	path = c.maskPath(app, group, tag, path)
	vl, err := c.store.GetValues([]string{path})
	if err != nil {
		log.Err(err).Msgf("获取多个配置项[%s]的配置信息出错:%+v", path, err)
	} else {
		log.Info().Msgf("获取多个配置项为:%+v", vl)
	}
	err = json.Unmarshal([]byte(vl[path]), clazz)
	return err
}

// Get 获取指定路径下的配置信息，并实现监听，当有数据变化时自动调用parser(配置数据的解析器，业务系统自定义实现)进行解析。
func (c configuration) Get(app, group, tag string, path []string, parser ChangedListener) {
	_path := make([]string, len(path))
	for i, v := range path {
		_path[i] = c.maskPath(app, group, tag, v)
	}
	vl, err := c.store.GetValues(_path)
	if err != nil {
		log.Err(err).Msgf("获取指定路径[%v]下的配置信息,并实现监听,当有数据变化时自动调用解析器进行解析出错:%+v", path, err)
	} else {
		log.Info().Msgf("获取多个配置项为:%v", vl)
	}
	parser.Changed(vl)
	go WatchProcessor(_path, c.store).Process(parser)
}

func (c configuration) Add(path string, value []byte) (string, error) {
	return c.store.Add(path, value)
}
