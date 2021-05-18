package config

import (
	"context"
	"encoding/json"
	"os"

	microConfig "github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source/etcd"

	"github.com/shelton-hu/logger"
)

const (
	// etcd config key = _BaseKeyPath + namespace + appName + (_SysConfPath or _CusConfPath)
	_BaseKeyPath = "/micro/config/"
	_SysConfPath = "/sysconf"
	_CusConfPath = "/cusconf"
)

var (
	// sysconf is the system config variable.
	sysconf *SystemConfig

	// cusconf is the custom config variable.
	cusconf *CustomConfig
)

// InitConfig initialize the system and custon config.
func InitConfig(ctx context.Context, etcdAddresses []string, namespace string, appName string) {
	initConfig(ctx, etcdAddresses, namespace, appName, _SysConfPath, sysconf)
	initConfig(ctx, etcdAddresses, namespace, appName, _CusConfPath, cusconf)
}

// SysConf returns the system config.
func SysConf() *SystemConfig {
	return sysconf
}

// CusConf returns the custom config.
func CusConf() *CustomConfig {
	return cusconf
}

// initConfig watch the config from etcd.
func initConfig(ctx context.Context, etcdAddresses []string, namespace string, appName string, etcdConfigPathSuffix string, conf interface{}) {
	source := etcd.NewSource(
		etcd.WithAddress(etcdAddresses...),
		etcd.WithPrefix(_BaseKeyPath+namespace+appName+_CusConfPath),
		etcd.StripPrefix(true),
	)

	mconf, err := microConfig.NewConfig()
	if err != nil {
		logger.Error(ctx, err.Error())
		os.Exit(1)
	}

	if err := mconf.Load(source); err != nil {
		logger.Error(ctx, "load config error: %s", err.Error())
	}

	go func(ctx context.Context, mconf microConfig.Config) {
		w, err := mconf.Watch()
		if err != nil {
			logger.Error(ctx, "config watch error: %s", err.Error())
			os.Exit(1)
		}
		for {
			v, err := w.Next()
			if err != nil {
				logger.Error(ctx, "watch next error，%s", err)
				os.Exit(1)
			}
			logger.Info(ctx, "system config was changed，%s", string(v.Bytes()))

			oldConfByte, _ := json.Marshal(conf)
			logger.Info(ctx, "system config old， %s", string(oldConfByte))
			err = json.Unmarshal(v.Bytes(), conf)
			if err != nil {
				logger.Error(ctx, "fatal error: change config value: %s", err.Error())
				os.Exit(1)
			}
			newConfByte, _ := json.Marshal(conf)
			logger.Info(ctx, "system config new，%s", string(newConfByte))
		}
	}(ctx, mconf)

	if err := mconf.Scan(conf); err != nil {
		logger.Error(ctx, "fatal error: scan config value: %s", err.Error())
		os.Exit(1)
	}
}
