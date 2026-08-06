package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestCastTimeCST 起卦时刻归一北京时间:无论进程时区(容器常为 UTC),
// castTime 返回的墙钟字段必须是东八区口径,且不改变绝对时刻。
func TestCastTimeCST(t *testing.T) {
	at, err := castTime(0) // 当下
	if err != nil {
		t.Fatal(err)
	}
	if _, off := at.Zone(); off != 8*3600 {
		t.Fatalf("castTime(0) 应归一东八区,得偏移 %d", off)
	}
	ts := time.Now().Add(-2 * time.Hour).Unix()
	at2, err := castTime(ts)
	if err != nil {
		t.Fatal(err)
	}
	if _, off := at2.Zone(); off != 8*3600 {
		t.Fatalf("castTime(unix) 应归一东八区,得偏移 %d", off)
	}
	if at2.Unix() != ts {
		t.Fatalf("归一时区不得改变绝对时刻: %d != %d", at2.Unix(), ts)
	}
}

// TestXiaoLiuRenBeijingHour 端到端钉住:UTC 进程起小六壬,时辰须按北京时间。
// 修复前(取进程本地墙钟)此测试在 UTC 环境必败(时辰错八小时)。
func TestXiaoLiuRenBeijingHour(t *testing.T) {
	ts := newTestServer(t, nil)
	now := time.Now()
	resp, raw := postJSON(t, ts.URL+"/api/v1/divination/xiaoliuren", map[string]any{"castAt": now.Unix()})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("起算应 200,得 %d: %s", resp.StatusCode, raw)
	}
	var env struct {
		Data struct {
			Result struct {
				LunarText string `json:"lunarText"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	branches := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	want := branches[((now.In(time.FixedZone("CST", 8*3600)).Hour()+1)/2)%12] + "时"
	if !strings.HasSuffix(env.Data.Result.LunarText, want) {
		t.Errorf("时辰应按北京时间为 %s,得 %q", want, env.Data.Result.LunarText)
	}
}

// TestValidateDivineReq 同一卦契约:castAt 必传;六爻须回传六掷;六壬不可代摇。
func TestValidateDivineReq(t *testing.T) {
	zero, one := 0, 1
	cases := []struct {
		name string
		req  divinationRequest
		code string // "" = 通过
	}{
		{"梅花缺castAt", divinationRequest{Kind: "meihua"}, "bad_cast_time"},
		{"梅花有castAt", divinationRequest{Kind: "meihua", CastAt: 1700000000}, ""},
		{"六爻缺tosses", divinationRequest{Kind: "liuyao", CastAt: 1700000000}, "bad_tosses"},
		{"六爻五掷不足", divinationRequest{Kind: "liuyao", CastAt: 1700000000, Tosses: []int{1, 2, 0, 3, 1}}, "bad_tosses"},
		{"六爻六掷齐备", divinationRequest{Kind: "liuyao", CastAt: 1700000000, Tosses: []int{1, 2, 0, 3, 1, 2}}, ""},
		{"六壬代摇拒绝", divinationRequest{Kind: "daliuren", CastAt: 1700000000, BaoShu: &zero}, "bad_baoshu"},
		{"六壬报数通过", divinationRequest{Kind: "daliuren", CastAt: 1700000000, BaoShu: &one}, ""},
		{"六壬正时通过", divinationRequest{Kind: "daliuren", CastAt: 1700000000}, ""},
	}
	for _, c := range cases {
		if code, _ := validateDivineReq(c.req); code != c.code {
			t.Errorf("%s: code=%q want %q", c.name, code, c.code)
		}
	}
}

// TestBodySizeLimit 超大请求体在解码层被拒(MaxBytesReader),不落入业务逻辑。
func TestBodySizeLimit(t *testing.T) {
	ts := newTestServer(t, nil)
	huge := map[string]any{"method": "time", "question": strings.Repeat("问", maxBodyBytes/3+1)}
	resp, _ := postJSON(t, ts.URL+"/api/v1/divination/meihua", huge)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("超限请求体应 400,得 %d", resp.StatusCode)
	}
	// 正常体量不受影响
	resp2, _ := postJSON(t, ts.URL+"/api/v1/divination/meihua", map[string]any{"method": "time", "question": "此番转职可成否"})
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("正常起卦应 200,得 %d", resp2.StatusCode)
	}
}
