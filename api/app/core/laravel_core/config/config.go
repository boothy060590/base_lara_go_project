package config_core

import (
	"base_lara_go_project/config"
)

// Re-export config package functions for Laravel-style access
var (
	Get                  = config.Get
	GetString            = config.GetString
	GetInt               = config.GetInt
	GetBool              = config.GetBool
	Has                  = config.Has
	Set                  = config.Set
	Load                 = config.Load
	ClearCache           = config.ClearCache
	Reload               = config.Reload
	ListAvailableConfigs = config.ListAvailableConfigs
	GetConfigLoader      = config.GetConfigLoader
)
