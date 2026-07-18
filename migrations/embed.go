// Package migrations 内嵌 SQL 迁移文件,由 internal/store 在启动时按序应用。
package migrations

import "embed"

// FS 迁移文件系统。
//
//go:embed *.sql
var FS embed.FS
