package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dawang20250107/ziweidoushu/internal/ai"
	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// chartRequest 排盘请求。
type chartRequest struct {
	Year          int      `json:"year"`
	Month         int      `json:"month"`
	Day           int      `json:"day"`
	Hour          int      `json:"hour"` // 时辰索引 0=早子时 ... 11=亥时 12=晚子时
	Gender        string   `json:"gender"`
	Name          string   `json:"name,omitempty"`
	Longitude     float64  `json:"longitude,omitempty"`
	Province      string   `json:"province,omitempty"`
	City          string   `json:"city,omitempty"`
	WorldCity     string   `json:"worldCity,omitempty"` // 国际出生地(查经度+时区)
	UTCOffset     *float64 `json:"utcOffset,omitempty"` // 直传时区偏移(小时);nil=未指定
	TrueSolarTime bool     `json:"trueSolarTime,omitempty"`
	ReferenceYear int      `json:"referenceYear,omitempty"`
	WithPatterns  *bool    `json:"withPatterns,omitempty"` // 默认 true
}

// chartResponse 排盘响应。
type chartResponse struct {
	Chart    *ziwei.Chart    `json:"chart"`
	Patterns []ziwei.Pattern `json:"patterns,omitempty"`
	Reading  *ai.Reading     `json:"reading,omitempty"` // 结构化多维断语(确定性,随盘生成)
}

// computeChart 排盘 + 格局(带 LRU 缓存)。
func (s *Server) computeChart(req chartRequest) (*chartResponse, error) {
	refYear := req.ReferenceYear
	if refYear == 0 {
		refYear = time.Now().Year()
	}
	longitude := req.Longitude
	var baseMeridian float64 // 0 → 引擎默认东经 120°(北京时)
	if req.TrueSolarTime {
		if req.WorldCity != "" && s.kb != nil { // 国际出生地:经度 + 时区标准经线
			if wc := s.kb.WorldCityOf(req.WorldCity); wc != nil {
				if longitude == 0 {
					longitude = wc.Longitude
				}
				baseMeridian = wc.UTCOffset * 15
			}
		} else if req.UTCOffset != nil {
			baseMeridian = *req.UTCOffset * 15
		}
		if longitude == 0 && s.kb != nil { // 国内省市回退
			longitude = s.kb.LongitudeOf(req.Province, req.City)
		}
	}
	withPatterns := req.WithPatterns == nil || *req.WithPatterns

	key := fmt.Sprintf("%d-%d-%d|%d|%s|%.2f|%.1f|%v|%d|%v",
		req.Year, req.Month, req.Day, req.Hour, req.Gender, longitude, baseMeridian, req.TrueSolarTime, refYear, withPatterns)
	if cached, ok := s.cache.Get(key); ok {
		s.metrics.chartCacheHit.Add(1)
		return cached.(*chartResponse), nil
	}
	s.metrics.chartCacheMiss.Add(1)

	chart, err := ziwei.Generate(ziwei.BirthInfo{
		Year: req.Year, Month: req.Month, Day: req.Day, Hour: req.Hour,
		Gender: ziwei.Gender(req.Gender), Name: req.Name,
		Longitude: longitude, BaseMeridian: baseMeridian,
	}, ziwei.Options{ReferenceYear: refYear, TrueSolarTime: req.TrueSolarTime})
	if err != nil {
		return nil, err
	}
	resp := &chartResponse{Chart: chart}
	if withPatterns {
		resp.Patterns = ziwei.DetectPatterns(chart)
	}
	if s.interp != nil { // 结构化多维断语随盘生成(不走 LLM,始终可用)
		resp.Reading = s.interp.BuildReading(chart, resp.Patterns)
	}
	s.cache.Set(key, resp)
	return resp, nil
}

func (s *Server) handleChart(w http.ResponseWriter, r *http.Request) {
	var req chartRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := s.computeChart(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_birth", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleWorldCities 世界主要城市(经度 + 时区),供国际出生地选择真太阳时。
func (s *Server) handleWorldCities(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"cities": s.kb.WorldCities})
}

func (s *Server) handleFamousList(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"persons": s.kb.Famous})
}

func (s *Server) handleFamousChart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, p := range s.kb.Famous {
		if p.ID != id {
			continue
		}
		resp, err := s.computeChart(chartRequest{
			Year: p.Year, Month: p.Month, Day: p.Day, Hour: p.Hour,
			Gender: p.Gender, Name: p.Name,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "chart_failed", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"person": p, "chart": resp.Chart, "patterns": resp.Patterns})
		return
	}
	writeError(w, http.StatusNotFound, "not_found", "未收录该名人")
}

// horoscopeRequest 运限请求:本命盘生辰 + 目标日期。
type horoscopeRequest struct {
	chartRequest
	Target struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Day   int `json:"day"`
		Hour  int `json:"hour"` // 时辰索引 0-12
	} `json:"target"`
}

// handleHoroscope 运限叠加:大限/小限/流年/流月/流日/流时。
// 倪师口径提示由引擎在字段注释中约定:生年四化与流年四化为准,
// 其余层 mutagen 为飞星派研究字段,前端展示层自行取舍。
func (s *Server) handleHoroscope(w http.ResponseWriter, r *http.Request) {
	var req horoscopeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := s.computeChart(req.chartRequest)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_birth", err.Error())
		return
	}
	h, err := ziwei.GenerateHoroscope(resp.Chart, req.Target.Year, req.Target.Month, req.Target.Day, req.Target.Hour)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_target", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"horoscope": h})
}

func (s *Server) handleLiuNianSiHua(w http.ResponseWriter, r *http.Request) {
	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if err != nil || year < 1900 || year > 2100 {
		writeError(w, http.StatusBadRequest, "bad_year", "year 参数需为 1900-2100 的整数")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"year":  year,
		"sihua": ziwei.LiuNianSiHua(year),
	})
}

func (s *Server) handleLiuYueSiHua(w http.ResponseWriter, r *http.Request) {
	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if err != nil || year < 1900 || year > 2100 {
		writeError(w, http.StatusBadRequest, "bad_year", "year 参数需为 1900-2100 的整数")
		return
	}
	month, err := strconv.Atoi(r.URL.Query().Get("month"))
	if err != nil || month < 1 || month > 12 {
		writeError(w, http.StatusBadRequest, "bad_month", "month 参数需为 1-12 的整数(农历月)")
		return
	}
	yearStem := ziwei.YearStemIndex(year)
	writeJSON(w, http.StatusOK, map[string]any{
		"year":  year,
		"month": month,
		"sihua": ziwei.LiuYueSiHua(yearStem, month),
	})
}
