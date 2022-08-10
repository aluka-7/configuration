package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/aluka-7/configuration"
	"github.com/aluka-7/utils"
)

/**
测试的配置文件为格式为：
{\"backend\":\"127.0.0.1:2181\",\"username\":\"guest\",\"password\":\"guest\"}
*/
func main() {
	// makeUaf()
	cfg := configuration.NewStoreConfig()
	//conf := configuration.Engine(cfg)
	p := &parser{}
	event := configuration.EventEngine(cfg)
	event.StartEventListener([]configuration.EventListener{p})
	for {
		var str string
		fmt.Scanln(&str)
		v, _ := configuration.NewFCEvent("test1")
		v.AddData("chirs", "zhou")
		event.Publish(v)
	}
}

func makeUaf() {
	key := configuration.DesKey
	src := "{\"backend\":\"common.zk.dev.local:2181\",\"username\":\"guest\",\"password\":\"guest\"}"
	fmt.Println("原文：" + src)
	enc, _ := utils.Encrypt([]byte(src), []byte(key))
	ioutil.WriteFile("configuration.uaf", []byte(base64.URLEncoding.EncodeToString(enc)), 0644)
	fmt.Println("密文：" + base64.URLEncoding.EncodeToString(enc))
	dec, _ := utils.Decrypt(enc, []byte(key))
	fmt.Println("解码：" + string(dec))
}

type parser struct {
}

var privileges = make(map[string][]string)

func (p parser) Changed(data map[string]string) {
	fmt.Printf("%+v\n", data)
	for _, v := range data {
		var vl map[string][]string
		if err := json.Unmarshal([]byte(v), &vl); err == nil {
			for k, _v := range vl {
				privileges[k] = _v
			}
		}
	}
}

/**
 * 该监听器要监听的事件key列表(可指定多个)，总是返回非空列表。
 *
 * @return
 */
func (p parser) EventKeys() []string {
	return []string{"chirs", "test"}
}

/**
 * 事件监听到的回调处理方法，由业务系统自行处理。
 *
 * @param event
 */
func (p parser) OnEvent(event configuration.Event) {
	fmt.Printf("OnEvent:%+v\n", event)
}
