package services_core

import (
	app_core "base_lara_go_project/app/core/app"
)

// Re-export interfaces from app_core for convenience
type BaseServiceInterface[T any] = app_core.BaseServiceInterface[T]
type CacheableServiceInterface[T any] = app_core.CacheableServiceInterface[T]
type SearchableServiceInterface[T any] = app_core.SearchableServiceInterface[T]
type AuditableServiceInterface[T any] = app_core.AuditableServiceInterface[T]
type AuditLog = app_core.AuditLog
type ServiceOptions = app_core.ServiceOptions
type ServiceFactory[T any] = app_core.ServiceFactory[T]
