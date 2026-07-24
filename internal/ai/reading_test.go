package ai

import (
	"strings"
	"testing"

	"github.com/dawang20250107/ziweidoushu/internal/ziwei"
)

// TestReadingDifferentiation 核心验收:不同命盘的多维断语必须显著不同,
// 且覆盖十二宫多维度(解决「不同人断语雷同」的问题)。
func TestReadingDifferentiation(t *testing.T) {
	a, err := ziwei.Generate(ziwei.BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 6, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2024})
	if err != nil {
		t.Fatal(err)
	}
	b, err := ziwei.Generate(ziwei.BirthInfo{Year: 1985, Month: 11, Day: 20, Hour: 0, Gender: ziwei.Female}, ziwei.Options{ReferenceYear: 2024})
	if err != nil {
		t.Fatal(err)
	}
	ra := buildReading(a, ziwei.DetectPatterns(a))
	rb := buildReading(b, ziwei.DetectPatterns(b))

	// 多维:至少 12 个维度(十二宫,大限视年龄)。
	if len(ra.Sections) < 12 {
		t.Fatalf("维度数应≥12,得 %d", len(ra.Sections))
	}
	// 十二宫每一维都须落地(防「交友/仆役」宫名与引擎不一致而被静默跳过)。
	wantKeys := []string{"ming", "caibo", "guanlu", "fuqi", "qianyi", "fude", "jie", "tianzhai", "zinv", "xiongdi", "jiaoyou", "fumu"}
	got := map[string]bool{}
	for _, s := range ra.Sections {
		got[s.Key] = true
	}
	for _, k := range wantKeys {
		if !got[k] {
			t.Errorf("缺失宫位维度 %s(检查宫名是否与引擎口径一致)", k)
		}
	}
	// 总论必须不同。
	if ra.Overview == rb.Overview {
		t.Errorf("两盘命格总论不应雷同")
	}
	// 逐宫比对:同名维度的断语大多应不同。
	byKeyB := map[string]string{}
	for _, s := range rb.Sections {
		byKeyB[s.Key] = s.Text
	}
	same, total := 0, 0
	for _, s := range ra.Sections {
		if tb, ok := byKeyB[s.Key]; ok {
			total++
			if tb == s.Text {
				same++
			}
		}
	}
	if total == 0 {
		t.Fatal("无可比对维度")
	}
	if same*100/total > 25 { // 允许极少数空宫巧合雷同,但绝大多数须不同
		t.Errorf("两盘断语雷同度过高:%d/%d 维度相同", same, total)
	}
}

// TestPairTraitOf 双主星组合定名与查断:与星序无关,须归一化到规范名。
func TestPairTraitOf(t *testing.T) {
	cases := []struct{ a, b, wantName string }{
		{"紫微", "破军", "紫微破军"},
		{"破军", "紫微", "紫微破军"}, // 反序须归一
		{"武曲", "贪狼", "武曲贪狼"},
		{"贪狼", "武曲", "武曲贪狼"},
		{"廉贞", "七杀", "廉贞七杀"},
		{"太阴", "太阳", "太阳太阴"},
	}
	for _, c := range cases {
		name, trait := pairTraitOf([]string{c.a, c.b})
		if name != c.wantName || trait == "" {
			t.Errorf("pairTraitOf(%s,%s)=%q(trait空=%v),期望规范名 %s 且有断语",
				c.a, c.b, name, trait == "", c.wantName)
		}
	}
	// 非组合(单星/三星/未知)须返回空。
	if n, _ := pairTraitOf([]string{"紫微"}); n != "" {
		t.Errorf("单主星不应有组合断语,得 %s", n)
	}
	if n, _ := pairTraitOf([]string{"紫微", "天府", "天相"}); n != "" {
		t.Errorf("三主星不应有双星组合断语,得 %s", n)
	}
	// 24 组组合断语键须全部可查且规范有序、无空文。
	for key, trait := range starPairTrait {
		if trait == "" {
			t.Errorf("组合 %s 断语为空", key)
		}
	}
	if len(starPairTrait) != 24 {
		t.Errorf("双主星组合应为 24 组,得 %d", len(starPairTrait))
	}
}

// TestSihuaLanding 生年四化落宫串联:四化须各自定位到宫,section 与总论贴盘。
func TestSihuaLanding(t *testing.T) {
	c, err := ziwei.Generate(ziwei.BirthInfo{Year: 1988, Month: 3, Day: 8, Hour: 4, Gender: ziwei.Female}, ziwei.Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}
	ls := sihuaLandings(c)
	if len(ls) == 0 {
		t.Fatal("应至少定位到部分生年四化落宫")
	}
	// 每个落宫记录须自洽:承化之星确实坐于该宫且带对应四化。
	for _, l := range ls {
		p := c.PalaceByName(l.Palace)
		if p == nil {
			t.Fatalf("四化落宫 %s 不存在", l.Palace)
		}
		st := p.FindStar(l.Star)
		if st == nil || st.SiHua != l.Hua {
			t.Errorf("%s化%s应确在【%s宫】且带该化", l.Star, l.Hua, l.Palace)
		}
	}
	rd := buildReading(c, ziwei.DetectPatterns(c))
	var sihuaText string
	for _, s := range rd.Sections {
		if s.Key == "sihua" {
			sihuaText = s.Text
		}
	}
	if !strings.Contains(sihuaText, "化忌入") && !strings.Contains(sihuaText, "化禄入") {
		t.Errorf("四化维度断语应含落宫串联,实际:%s", sihuaText)
	}
	// 总论应点出化忌坐宫(若盘中有化忌)。
	hasJi := false
	for _, l := range ls {
		if l.Hua == ziwei.HuaJi {
			hasJi = true
		}
	}
	if hasJi && !strings.Contains(rd.Overview, "化忌坐") {
		t.Errorf("总论应点出生年化忌坐宫,实际:%s", rd.Overview)
	}
}

// TestLiuNianSection 流年维度:流年命宫落宫 + 流年四化飞入本命宫位,须贴盘且随年而变。
func TestLiuNianSection(t *testing.T) {
	c26, err := ziwei.Generate(ziwei.BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 6, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}
	c27, err := ziwei.Generate(ziwei.BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 6, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2027})
	if err != nil {
		t.Fatal(err)
	}
	get := func(c *ziwei.Chart) ReadingSection {
		for _, s := range buildReading(c, ziwei.DetectPatterns(c)).Sections {
			if s.Key == "liunian" {
				return s
			}
		}
		t.Fatal("缺失流年维度")
		return ReadingSection{}
	}
	s26, s27 := get(c26), get(c27)
	// 干支须随年而变(2026 丙午、2027 丁未)。
	if !strings.Contains(s26.Title, "2026") || !strings.Contains(s27.Title, "2027") {
		t.Errorf("流年标题应含年份,得 %q / %q", s26.Title, s27.Title)
	}
	// 同一本命盘、不同流年,断语必须不同(流年命宫与四化都不同)。
	if s26.Text == s27.Text {
		t.Error("不同流年的断语不应雷同")
	}
	// 须含流年四化飞入本命宫位的表述。
	if !strings.Contains(s26.Text, "飞入本命") {
		t.Errorf("流年断语应含四化飞入本命宫位,实际:%s", s26.Text)
	}
}

// TestStarPalaceCompleteness 宫位透镜按星细化:14×12 须齐备、无空、且随宫而异。
func TestStarPalaceCompleteness(t *testing.T) {
	stars := []string{"紫微", "天机", "太阳", "武曲", "天同", "廉贞", "天府", "太阴", "贪狼", "巨门", "天相", "天梁", "七杀", "破军"}
	palaces := []string{"命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄", "迁移", "仆役", "官禄", "田宅", "福德", "父母"}
	if len(starPalaceTrait) != 14 {
		t.Errorf("主星数应为 14,得 %d", len(starPalaceTrait))
	}
	for _, s := range stars {
		m, ok := starPalaceTrait[s]
		if !ok {
			t.Errorf("缺主星 %s", s)
			continue
		}
		if len(m) != 12 {
			t.Errorf("%s 宫位数应为 12,得 %d", s, len(m))
		}
		for _, p := range palaces {
			if strings.TrimSpace(m[p]) == "" {
				t.Errorf("%s×%s 断语为空", s, p)
			}
		}
	}
	// 同一星在不同宫读法须不同(准头的意义所在)。
	if starInPalace("太阴", "财帛") == starInPalace("太阴", "夫妻") {
		t.Error("太阴在财帛与夫妻断语不应相同")
	}
	if starInPalace("太阳", "官禄") == starInPalace("太阳", "疾厄") {
		t.Error("太阳在官禄与疾厄断语不应相同")
	}
}

// TestEscalationFires 应验分档:吉档(文贵/财禄)与凶档(血光/冲击/是非/损耗/成空)
// 均须能在样本盘中触发,证明煞吉会照确实推进到应验档位。
func TestEscalationFires(t *testing.T) {
	goodHit, badHit := false, false
	badMarkers := []string{"血光档", "冲击档", "是非档", "损耗档", "成空档"}
	goodMarkers := []string{"文贵档", "财禄档", "暴发格"}
	for y := 1975; y <= 1995 && !(goodHit && badHit); y++ {
		for mo := 1; mo <= 12 && !(goodHit && badHit); mo++ {
			for _, g := range []ziwei.Gender{ziwei.Male, ziwei.Female} {
				for h := 0; h < 12; h += 3 {
					c, err := ziwei.Generate(ziwei.BirthInfo{Year: y, Month: mo, Day: 8, Hour: h, Gender: g}, ziwei.Options{ReferenceYear: 2026})
					if err != nil {
						continue
					}
					for _, s := range buildReading(c, ziwei.DetectPatterns(c)).Sections {
						for _, mk := range badMarkers {
							if strings.Contains(s.Text, mk) {
								badHit = true
							}
						}
						for _, mk := range goodMarkers {
							if strings.Contains(s.Text, mk) {
								goodHit = true
							}
						}
					}
				}
			}
		}
	}
	if !goodHit {
		t.Error("样本盘中未触发任何吉向应验档(文贵/财禄/暴发)")
	}
	if !badHit {
		t.Error("样本盘中未触发任何凶向应验档(血光/冲击/是非/损耗/成空)")
	}
}

// TestHealthLayer 疾厄宫健康专层:健康专断 + 免责声明,且按星曜给身体部位。
func TestHealthLayer(t *testing.T) {
	// 遍历样本,取到有实配主星的疾厄宫,验证健康专断落地。
	for y := 1985; y <= 1995; y++ {
		for h := 0; h < 12; h += 2 {
			c, err := ziwei.Generate(ziwei.BirthInfo{Year: y, Month: 6, Day: 12, Hour: h, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2026})
			if err != nil {
				continue
			}
			jie := c.PalaceByName("疾厄")
			if jie == nil || len(jie.MajorStarNames()) == 0 {
				continue
			}
			var text string
			for _, s := range buildReading(c, ziwei.DetectPatterns(c)).Sections {
				if s.Key == "jie" {
					text = s.Text
				}
			}
			if !strings.Contains(text, "健康专断") {
				t.Fatalf("疾厄有主星却无健康专断:%s", text)
			}
			if !strings.Contains(text, "不代医疗诊断") {
				t.Errorf("健康断语应含免责声明:%s", text)
			}
			return // 验证一例即可
		}
	}
}

// TestAnnualTiming 流年择时:健康预警与财官催旺两维度须随盘给出,择时年落十年窗口。
func TestAnnualTiming(t *testing.T) {
	c, err := ziwei.Generate(ziwei.BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 6, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2026})
	if err != nil {
		t.Fatal(err)
	}
	secs := buildReading(c, ziwei.DetectPatterns(c)).Sections
	got := map[string]string{}
	for _, s := range secs {
		got[s.Key] = s.Text
	}
	if _, ok := got["healthtiming"]; !ok {
		t.Error("缺健康预警·流年维度")
	}
	ft, ok := got["fortunetiming"]
	if !ok {
		t.Error("缺财官择时·流年维度")
	}
	if !strings.Contains(ft, "财运") || !strings.Contains(ft, "事业") {
		t.Errorf("财官择时应含财运与事业两段:%s", ft)
	}
	// 择日下钻:本年月度择时维度须给出。
	mt, ok := got["monthtiming"]
	if !ok || !strings.Contains(mt, "月度择时") {
		t.Errorf("缺本年月度择时维度:%q", mt)
	}
	// 择时年份须落在 ReferenceYear 起十年窗口。
	health, wealth, career := annualTiming(c)
	for _, grp := range [][]TimingYear{health, wealth, career} {
		for _, ty := range grp {
			if ty.Year < 2026 || ty.Year >= 2036 {
				t.Errorf("择时年 %d 越出十年窗口", ty.Year)
			}
			if ty.GanZhi == "" || ty.Note == "" {
				t.Errorf("择时年 %d 缺干支或理由", ty.Year)
			}
		}
	}
}

// TestReadingChartGrounded 断语须引用本盘实配星曜(非通用套话)。
func TestReadingChartGrounded(t *testing.T) {
	c, err := ziwei.Generate(ziwei.BirthInfo{Year: 1990, Month: 6, Day: 15, Hour: 6, Gender: ziwei.Male}, ziwei.Options{ReferenceYear: 2024})
	if err != nil {
		t.Fatal(err)
	}
	rd := buildReading(c, ziwei.DetectPatterns(c))
	// 命宫维度的断语应含命宫的某颗主星名(或借星)。
	ming := c.MingGong()
	stars := ming.MajorStarNames()
	if len(stars) == 0 {
		stars = ming.BorrowedStars
	}
	if len(stars) == 0 {
		return // 极少数命宫空且无借星,跳过
	}
	var mingText string
	for _, s := range rd.Sections {
		if s.Key == "ming" {
			mingText = s.Text
		}
	}
	hit := false
	for _, name := range stars {
		if strings.Contains(mingText, name) {
			hit = true
		}
	}
	if !hit {
		t.Errorf("命宫断语应引用命宫主星%v,实际:%s", stars, mingText)
	}
	// 每个维度都应有 level 且文本非空。
	for _, s := range rd.Sections {
		if s.Text == "" || s.Level == "" {
			t.Errorf("维度 %s 断语或吉凶档缺失", s.Key)
		}
	}
}
