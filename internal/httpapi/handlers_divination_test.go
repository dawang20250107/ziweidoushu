package httpapi

import (
	"net/http"
	"strings"
	"testing"
)

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
