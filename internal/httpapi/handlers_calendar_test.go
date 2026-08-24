package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// TestLunarCalendarEndpoints 农历端点:年月表(闰月/大小月)+ 换算(实测锚点)+ 拒绝路径。
func TestLunarCalendarEndpoints(t *testing.T) {
	ts := newTestServer(t, nil)
	defer ts.Close()

	// 年月表:1993 有闰三月 → 13 个月
	resp, err := http.Get(ts.URL + "/api/v1/calendar/lunar-year?year=1993")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var yearBody struct {
		OK   bool `json:"ok"`
		Data struct {
			Year   int `json:"year"`
			Months []struct {
				Month int  `json:"month"`
				Leap  bool `json:"leap"`
				Days  int  `json:"days"`
			} `json:"months"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&yearBody); err != nil {
		t.Fatal(err)
	}
	if !yearBody.OK || len(yearBody.Data.Months) != 13 {
		t.Fatalf("1993 应 13 个月(含闰三月),得 %d", len(yearBody.Data.Months))
	}
	hasLeap3 := false
	for _, m := range yearBody.Data.Months {
		if m.Leap && m.Month == 3 {
			hasLeap3 = true
		}
	}
	if !hasLeap3 {
		t.Fatal("1993 月表应含闰三月")
	}

	// 换算:农历一九九四年正月十五 → 公历 1994-02-24(实测案例)
	conv := func(payload string) (int, map[string]any) {
		r, err := http.Post(ts.URL+"/api/v1/calendar/lunar-to-solar", "application/json", bytes.NewBufferString(payload))
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()
		var body struct {
			OK   bool           `json:"ok"`
			Data map[string]any `json:"data"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		return r.StatusCode, body.Data
	}
	code, data := conv(`{"year":1994,"month":1,"leap":false,"day":15}`)
	if code != 200 || data["year"].(float64) != 1994 || data["month"].(float64) != 2 || data["day"].(float64) != 24 {
		t.Fatalf("1994 正月十五应换算为 1994-2-24,得 %d %v", code, data)
	}
	// 闰月锚点:2020 闰四月初一 → 2020-05-23
	code, data = conv(`{"year":2020,"month":4,"leap":true,"day":1}`)
	if code != 200 || data["month"].(float64) != 5 || data["day"].(float64) != 23 {
		t.Fatalf("2020 闰四月初一应换算为 2020-5-23,得 %d %v", code, data)
	}
	// 拒绝:1994 无闰正月
	if code, _ = conv(`{"year":1994,"month":1,"leap":true,"day":1}`); code != 400 {
		t.Fatalf("不存在的闰月应 400,得 %d", code)
	}
	// 拒绝:区间外
	if code, _ = conv(`{"year":1899,"month":1,"leap":false,"day":1}`); code != 400 {
		t.Fatalf("区间外年份应 400,得 %d", code)
	}
}
