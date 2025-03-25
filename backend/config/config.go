// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"

	"github.com/CloudDetail/metadata/configs"
	"github.com/spf13/viper"
)

var config *Config

type Config struct {
	Server struct {
		Port                     int `mapstructure:"port"`
		AccessTokenExpireMinutes int `mapstructure:"access_token_expire_minutes"`
		RefreshTokenExpireHours  int `mapstructure:"refresh_token_expire_hours"`
	} `mapstructure:"server"`
	Logger struct {
		Level         string `mapstructure:"level"`
		EnableConsole bool   `mapstructure:"console_enable"`
		EnableFile    bool   `mapstructure:"file_enable"`
		FilePath      string `mapstructure:"file_path"`
		FileNum       int    `mapstructure:"file_num"`
		FileSize      int    `mapstructure:"file_size_mb"`
	} `mapstructure:"logger"`
	Database struct {
		Connection string `mapstructure:"connection"`
		MaxOpen    int    `mapstructure:"max_open"`
		MaxIdle    int    `mapstructure:"max_idle"`
		MaxLife    int    `mapstructure:"max_life_second"`
		MySql      struct {
			Host     string `mapstructure:"host"`
			Port     int    `mapstructure:"port"`
			Database string `mapstructure:"database"`
			UserName string `mapstructure:"username"`
			Password string `mapstructure:"password"`
			Charset  string `mapstructure:"charset"`
		} `mapstructure:"mysql"`
		Sqllite struct {
			Database string `mapstructure:"database"`
		} `mapstructure:"sqllite"`
		Postgres struct {
			Host     string `mapstructure:"host"`
			Port     int    `mapstructure:"port"`
			Database string `mapstructure:"database"`
			UserName string `mapstructure:"username"`
			Password string `mapstructure:"password"`
			SSLMode  string `mapstructure:"sslmode"`
			Timezone string `mapstructure:"timezone"`
		}
	} `mapstructure:"database"`
	ClickHouse struct {
		Address  string `mapstructure:"address"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Database string `mapstructure:"database"`
		Cluster  string `mapstructure:"cluster"`
		Replica  bool   `mapstructure:"replica"`
	} `mapstructure:"clickhouse"`
	Promethues struct {
		Address string `mapstructure:"address"`
		Storage string `mapstructure:"storage"`
	} `mapstructure:"promethues"`
	Kubernetes struct {
		AuthType     string `mapstructure:"auth_type"`
		AuthFilePath string `mapstructure:"auth_file_path"`

		MetadataSettings MetadataSettings `mapstructure:"metadata_settings"`
	} `mapstructure:"kubernetes"`
	MetaServer struct {
		Enable           bool                     `mapstructure:"enable"`
		MetaSourceConfig configs.MetaSourceConfig `mapstructure:"meta_source_config"`
	} `mapstructure:"meta_server"`
	Jaeger struct {
		Address string `mapstructure:"address"`
	}
	DeepFlow struct {
		ServerAddress string `mapstructure:"server_address"`
		ChAddress     string `mapstructure:"ch_address"`
		ChUsername    string `mapstructure:"ch_username"`
		ChPassword    string `mapstructure:"ch_password"`
	} `mapstructure:"deepflow"`
	User struct {
		AnonymousUser `mapstructure:"anonymous_user"`
	} `mapstructure:"user"`

	Dify struct {
		User    string `mapstructure:"user"`
		URL     string `mapstructure:"url"`
		APIKeys struct {
			AlertCheck string `mapstructure:"alert_check"`
		} `mapstructure:"api_keys"`
		FlowIDs struct {
			AlertCheck        string `mapstructure:"alert_check"`
			AlertEventAnalyze string `mapstructure:"alert_event_analyze"`
		} `mapstructure:"flow_ids"`
		MaxConcurrency int    `mapstructure:"max_concurrency"`
		CacheMinutes   int    `mapstructure:"cache_minutes"`
		Sampling       string `mapstructure:"sampling"`
	} `mapstructure:"dify"`
}

type AnonymousUser struct {
	Username string `mapstructure:"username"`
	Enable   bool   `mapstructure:"enable"`
	Role     string `mapstructure:"role"`
}

type MetadataSettings struct {
	Namespace         string `mapstructure:"namespace"`
	AlertRuleCMName   string `mapstructure:"alert_rule_configmap_name"`
	AlertRuleFileName string `mapstructure:"alert_rule_file_name"`

	AlertManagerCMName   string `mapstructure:"alert_manager_configmap"`
	AlertManagerFileName string `mapstructure:"alert_manager_file_name"`
	VectorCMName         string `mapstructure:"vector_configmap"`
	VectorFileName       string `mapstructure:"vector_file_name"`
}

func Get() *Config {
	if config == nil {
		viper.SetConfigType("yaml")
		configFile, found := os.LookupEnv("APO_CONFIG")
		if !found {
			configFile = "./config/apo.yml"
		}
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
		if err := viper.Unmarshal(&config); err != nil {
			panic(err)
		}
	}
	return config
}

func GetCHCluster() string {
	return config.ClickHouse.Cluster
}
