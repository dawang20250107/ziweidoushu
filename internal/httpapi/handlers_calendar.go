// 农历日历端点:农历年月表 + 农历→公历换算(排盘表单的农历输入模式)。
// 换算在提交前完成,下游(排盘/档案/问星)仍以公历为唯一事实源。
package httpapi

import (
	"net/http"
	"strconv"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// handleLunarYear GET /api/v1/calendar/lunar-year?year=1993
// → { year, months: [{month, leap, days}...] }(闰月按年内实际位置插入)
func (s *Server) handleLunarYear(w http.ResponseWriter, r *http.Request) {
	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_year", "year 参数须为整数")
		return
	}
	months, err := ziwei.LunarYearMonths(year)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_year", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"year": year, "months": months})
}

type lunarToSolarRequest struct {
	Year  int  `json:"year"`
	Month int  `json:"month"`
	Leap  bool `json:"leap"`
	Day   int  `json:"day"`
}

// handleLunarToSolar POST /api/v1/calendar/lunar-to-solar
// {year, month, leap, day} → {year, month, day}(公历)
func (s *Server) handleLunarToSolar(w http.ResponseWriter, r *http.Request) {
	var req lunarToSolarRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	sy, sm, sd, err := ziwei.LunarToSolar(req.Year, req.Month, req.Leap, req.Day)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_lunar_date", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"year": sy, "month": sm, "day": sd})
}
