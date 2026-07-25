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
			"school":   "倪海厦《天纪》三合派(生年四化固定)",
			"parity":   "iztro 2.5.8(《紫微斗数全书》安星法,黄金基准 1500+ 用例回归)",
			"calendar": "lunar-go(1900-2100)",
			// 历法口径:各术数线分界规则(两派并存处均有测试钉死)
			"calendarCaliber": map[string]string{
				"ziwei":    "年界正月初一、月按农历初一分界(五虎遁,闰月过半折下月)、日柱 23 点归次日",
				"sizhu":    "子平节气口径:年起立春、月起节(精确到交接时刻)、晚子时日柱归次日",
				"liuyao":   "月建以节交接时刻分界,日辰 23 点归次日(与四柱同)",
				"daliuren": "月将按最近中气换将(雨水亥将逐支退),日干支 23 点归次日;占时正时/活时报数两制(活时自子顺数)",
				"meihua":   "农历月日取数(闰月按当月数),民用日 0 点分界(掐指口径);有问辞按字数起数(声音占义),无问辞守年月日时",
			},
		},
		"ai": map[string]any{
			"provider": s.interp.ProviderName(),
		},
		"corpus":    s.corpus.Stats(),
		"goVersion": runtime.Version(),
	})
}
