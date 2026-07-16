// Package data 内置数据资产(古籍原文、倪海厦三纪知识库、合盘知识、城市经度、名人盘)。
//
// 全部为 JSON 静态打包:服务启动即加载进内存,零数据库依赖。
// 后续新增古籍/倪师著作既可提交到本目录(随二进制打包),
// 也可放到外部目录由 CORPUS_EXTERNAL_DIR 热加载(见 internal/corpus)。
package data

import "embed"

// FS 打包后的只读数据文件系统。
//
//go:embed classics/*.json nihai/*.json knowledge/*.json cities.json famous.json
var FS embed.FS
