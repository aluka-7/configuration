package configuration_test

import (
	"fmt"
	"testing"

	"github.com/aluka-7/configuration"
	"github.com/aluka-7/configuration/backends"
)

func TestString(t *testing.T) {
	conf := configuration.MockEngine(t, backends.StoreConfig{Exp: map[string]string{
		"/system/base/cache/provider":  "{\"test\":\"test\"}",
		"/system/base/rpc/client/1000": "system.manage.svc:9191",
	}})
	actual, _ := conf.String("base", "cache", "", "provider")

	expected := "{\"test\":\"test\"}"
	if actual != expected {
		t.Error("生成的结果不匹配\n", "预期:", expected, "|", "实际:", actual)
	}

	actual, _ = conf.String("base", "rpc", "client", "1000")
	expected = "system.manage.svc:9191"
	if actual != expected {
		t.Error("生成的结果不匹配\n", "预期:", expected, "|", "实际:", actual)
	}
}

func TestValues(t *testing.T) {
	conf := configuration.MockEngine(t, backends.StoreConfig{Exp: map[string]string{
		"/system/base/rpc/client/base": "{\"test\":\"test\"}",
		"/system/base/rpc/client/1000": "system.manage.svc:9191",
	}})

	actual, _ := conf.Values("base", "rpc", "client", []string{"base", "1000"})
	for k, v := range actual {
		expected := ""
		switch k {
		case "/system/base/rpc/client/base":
			expected = "{\"test\":\"test\"}"
		case "/system/base/rpc/client/1000":
			expected = "system.manage.svc:9191"
		default:
			t.Error("没有获取到值")
		}
		if v != expected {
			t.Error("生成的结果不匹配\n", "预期:", expected, "|", "实际:", v)
		}
	}
}

type test struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func TestClazz(t *testing.T) {
	conf := configuration.MockEngine(t, backends.StoreConfig{Exp: map[string]string{
		"/system/base/rpc/client/base": "{\"key\":\"client\",\"value\":\"base\"}",
		"/system/base/rpc/client/1000": "{\"key\":\"client\",\"value\":\"1000\"}",
	}})
	var ts test
	conf.Clazz("base", "rpc", "client", "base", &ts)
	expected := "{Key:client Value:base}"
	actual := fmt.Sprintf("%+v", ts)
	if actual != expected {
		t.Error("生成的结果不匹配\n", "预期:", expected, "|", "实际:", actual)
	}

	conf.Clazz("base", "rpc", "client", "1000", &ts)
	expected = "{Key:client Value:1000}"
	actual = fmt.Sprintf("%+v", ts)
	if actual != expected {
		t.Error("生成的结果不匹配\n", "预期:", expected, "|", "实际:", actual)
	}
}
