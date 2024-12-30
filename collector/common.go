package collector

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
)

func GetHttpConfig() *config.HttpConfig {
	httpConfig := config.DefaultHttpConfig()
	if !isProxyValid() {
		return httpConfig
	}

	global := CloudConf.Global
	proxy := config.Proxy{
		Schema: global.HttpSchema,
		Host:   global.HttpHost,
		Port:   global.HttpPort,
	}
	if isUserInfoValid() {
		if ProxyEnabled {
			proxy.Username = TmpProxyUserName
			proxy.Password = TmpProxyPassword
		} else {
			proxy.Username = global.UserName
			proxy.Password = global.Password
		}
	}
	httpConfig.HttpProxy = &proxy
	return httpConfig
}

func isProxyValid() bool {
	global := CloudConf.Global
	return global.HttpSchema != "" && global.HttpHost != "" && global.HttpPort > 0
}

func isUserInfoValid() bool {
	global := CloudConf.Global
	return global.UserName != "" && global.Password != ""
}
