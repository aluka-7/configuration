# Aluka-7 配置管理引擎

提供给所有业务系统使用的配置管理引擎(基于Zookeeper)，所有业务系统/中间件的可变配置通过集中配置中心进行统一配置，业务系统可通过该类来管理自己的配置数据并在配置中心的数据发生变化时得到及时的通知。

配置管理中心的数据存储为树形目录结构（类似文件系统的目录结构），每一个节点都可以存储相应的数据。业务系统可在基础目录树的基础上往下继续扩展新的目录结构，从而可以区分存储不同的配置项信息。比如系统分配给业务系统A的基础目录（称为path）为 `/sysa `，那么业务系统可以扩充到如下path： `/sysa/key1、/sysa/key2、/sysa/key3 `，每个path可存储对应的数据。

配置管理引擎客户端依赖连接配置管理中心，需要在同一配置管理中心下载对应的授权文件方可。


## 快速使用

1. 获取配置管理引擎的唯一实例

```go

config:=configuration.Engine()
```

2. 获取多个配置项的配置信息，返回原始的配置数据格式(map集合)，如果获取失败则抛出异常。

```go
config.Values(path []string) (map[string]string, error)
```

3. 获取指定配置项的配置信息，返回原始的配置数据格式，如果获取失败则抛出异常。

```go
config.String(path string) (string, error) 
```

4. 获取指定配置项的配置信息，并且将配置信息（JSON格式的）转换为指定的Go结构体，如果获取失败或转换失败则抛出异常。

```go
config.Clazz(path string, clazz interface{}) error 
```

5. 判断些path是否存在
```go
config.Has(path string) (bool, error)
```

6. 获取指定路径下的配置信息，并实现监听，当有数据变化时自动调用解析器进行解析。

```go
par := &Parser{path: []string{path}}
config.Get(path []string, par).Process([]string{path}, par)
```

```go
type Parser struct {
    path []string
}

func (p Parser) Parse(data map[string]string) {
   //此处对于获取的值进行处理,map的key为path
}
func (p Parser) Changed(data map[string]string) {
    p.Parse(data)
}
```
