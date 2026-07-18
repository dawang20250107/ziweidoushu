package httpapi

import (
	"net/http"
	"runtime"
)

// handleMeta 服务元信息:版本、引擎口径、AI 供应商、语料统计。
func (s *Server) handleMeta(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    "ziweidoushu",
		"version": Version,
		"engine": map[string]any{
			"school":     "倪海厦《天纪》三合派(生年四化固定)",
			"parity":     "iztro 2.5.8(《紫微斗数全书》安星法,黄金基准 1500+ 用例回归)",
			"calendar":   "lunar-go(1900-2100)",
			"yearDivide": "正月初一",
		},
		"ai": map[string]any{
			"provider": s.interp.ProviderName(),
		},
		"corpus":    s.corpus.Stats(),
		"goVersion": runtime.Version(),
	})
}
