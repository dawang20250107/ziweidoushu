// Package ziwei 实现紫微斗数排盘引擎。
//
// 体系口径:倪海厦《天纪》三合派正统——生年四化固定不动,
// 不使用飞星派宫干自化/大限四化作为排盘输出(工具函数保留供研究)。
// 安星算法与 iztro 2.5.8 逐字段对齐,由 1500+ 黄金基准用例回归保障。
package ziwei

// Gender 性别。
type Gender string

const (
	Male   Gender = "male"
	Female Gender = "female"
)

// BirthInfo 排盘输入(公历)。
type BirthInfo struct {
	Year   int    `json:"year"`
	Month  int    `json:"month"` // 1-12
	Day    int    `json:"day"`
	Hour   int    `json:"hour"` // 时辰地支索引 0=子 ... 11=亥
	Gender Gender `json:"gender"`
	Name   string `json:"name,omitempty"`
	// Longitude 出生地东经度数;非零时按真太阳时校正时辰(以东经 120° 为基准)。
	Longitude float64 `json:"longitude,omitempty"`
}

// LunarInfo 农历信息。
type LunarInfo struct {
	LunarYear   int  `json:"lunarYear"`
	LunarMonth  int  `json:"lunarMonth"` // 1-12(闰月时 IsLeapMonth=true)
	LunarDay    int  `json:"lunarDay"`
	YearStem    int  `json:"yearStem"`   // 0=甲 ... 9=癸
	YearBranch  int  `json:"yearBranch"` // 0=子 ... 11=亥
	IsLeapMonth bool `json:"isLeapMonth"`
}

// SiHua 四化类型。
type SiHua string

const (
	HuaLu   SiHua = "禄"
	HuaQuan SiHua = "权"
	HuaKe   SiHua = "科"
	HuaJi   SiHua = "忌"
)

// StarType 星曜分类。
type StarType string

const (
	StarMajor StarType = "major" // 十四主星
	StarLucky StarType = "lucky" // 吉星(六吉、禄存天马等)
	StarSha   StarType = "sha"   // 煞星(六煞等)
	StarMinor StarType = "minor" // 其余杂曜
)

// Star 一颗安入宫位的星曜。
type Star struct {
	Name string   `json:"name"`
	Type StarType `json:"type"`
	// Brightness 庙旺得利平不陷 原始亮度(部分星曜无亮度,为空)。
	Brightness string `json:"brightness,omitempty"`
	// BrightnessLevel 三档简化:bright(庙旺) / normal(得利平) / dim(不陷)。
	BrightnessLevel string `json:"brightnessLevel,omitempty"`
	// SiHua 生年四化标记(禄/权/科/忌),无则为空。
	SiHua SiHua `json:"siHua,omitempty"`
}

// Palace 十二宫之一。
type Palace struct {
	Branch int    `json:"branch"` // 地支索引 0=子 ... 11=亥
	Stem   int    `json:"stem"`   // 天干索引 0=甲 ... 9=癸
	Name   string `json:"name"`   // 宫名(命宫/兄弟/夫妻/子女/财帛/疾厄/迁移/仆役/官禄/田宅/福德/父母)
	Stars  []Star `json:"stars"`

	DaXianStart int `json:"daXianStart"` // 大限起始虚岁
	DaXianEnd   int `json:"daXianEnd"`   // 大限结束虚岁

	IsMingGong bool `json:"isMingGong"`
	IsShenGong bool `json:"isShenGong"`

	// OppositeBranch 对宫地支索引 = (Branch+6)%12。
	OppositeBranch int `json:"oppositeBranch"`
	// IsEmpty 空宫(无主星)。
	IsEmpty bool `json:"isEmpty"`
	// 空宫借对宫主星(结构化,供文案层直接使用)。
	BorrowedFromBranch int      `json:"borrowedFromBranch,omitempty"`
	BorrowedFromName   string   `json:"borrowedFromName,omitempty"`
	BorrowedStars      []string `json:"borrowedStars,omitempty"`

	// Changsheng12 长生十二神(长生/沐浴/冠带/临官/帝旺/衰/病/死/墓/绝/胎/养)。
	Changsheng12 string `json:"changsheng12,omitempty"`
	// Boshi12 博士十二神。
	Boshi12 string `json:"boshi12,omitempty"`
	// Ages 小限虚岁序列。
	Ages []int `json:"ages,omitempty"`
}

// DaXian 大限(十年运)。倪师体系:四化永远固定,大限只看宫位移动。
type DaXian struct {
	StartAge     int    `json:"startAge"` // 虚岁
	EndAge       int    `json:"endAge"`
	PalaceBranch int    `json:"palaceBranch"`
	PalaceName   string `json:"palaceName"`
}

// FourPillars 四柱干支(年月日时)。
type FourPillars struct {
	Year  string `json:"year"`
	Month string `json:"month"`
	Day   string `json:"day"`
	Hour  string `json:"hour"`
}

// Chart 完整命盘。
type Chart struct {
	BirthInfo BirthInfo `json:"birthInfo"`
	LunarInfo LunarInfo `json:"lunarInfo"`
	// LunarDateText 农历中文表示,如「一九九〇年五月廿三」。
	LunarDateText string      `json:"lunarDateText"`
	FourPillars   FourPillars `json:"fourPillars"`
	TimeName      string      `json:"timeName"` // 如「午时」
	Zodiac        string      `json:"zodiac"`   // 生肖
	Sign          string      `json:"sign"`     // 星座

	MingGongBranch int    `json:"mingGongBranch"`
	ShenGongBranch int    `json:"shenGongBranch"`
	MingZhu        string `json:"mingZhu"`      // 命主星
	ShenZhu        string `json:"shenZhu"`      // 身主星
	WuxingJu       int    `json:"wuxingJu"`     // 2/3/4/5/6
	WuxingJuName   string `json:"wuxingJuName"` // 如「水二局」
	ZiweiPos       int    `json:"ziweiPos"`     // 紫微星地支索引

	// Palaces 十二宫,按地支索引 0-11 排序(索引即地支)。
	Palaces []Palace `json:"palaces"`
	DaXians []DaXian `json:"daXians"`

	// CurrentAge 按 ReferenceYear 计算的虚岁近似(参考年-出生年);
	// CurrentDaXianIndex 当前大限序号,无匹配为 -1。
	ReferenceYear      int `json:"referenceYear"`
	CurrentAge         int `json:"currentAge"`
	CurrentDaXianIndex int `json:"currentDaXianIndex"`
}

// PalaceByBranch 按地支索引取宫。
func (c *Chart) PalaceByBranch(branch int) *Palace {
	b := ((branch % 12) + 12) % 12
	for i := range c.Palaces {
		if c.Palaces[i].Branch == b {
			return &c.Palaces[i]
		}
	}
	return nil
}

// PalaceByName 按宫名取宫(iztro 口径宫名,如「夫妻」「仆役」)。
func (c *Chart) PalaceByName(name string) *Palace {
	for i := range c.Palaces {
		if c.Palaces[i].Name == name {
			return &c.Palaces[i]
		}
	}
	return nil
}

// MingGong 命宫。
func (c *Chart) MingGong() *Palace { return c.PalaceByBranch(c.MingGongBranch) }

// SanFangSiZheng 三方四正:命宫+官禄+财帛+迁移(对宫)。
func (c *Chart) SanFangSiZheng() []*Palace {
	m := c.MingGongBranch
	out := make([]*Palace, 0, 4)
	for _, b := range []int{m, (m + 4) % 12, (m + 8) % 12, (m + 6) % 12} {
		if p := c.PalaceByBranch(b); p != nil {
			out = append(out, p)
		}
	}
	return out
}

// HasStar 宫内是否有某星。
func (p *Palace) HasStar(name string) bool {
	for i := range p.Stars {
		if p.Stars[i].Name == name {
			return true
		}
	}
	return false
}

// FindStar 取宫内某星,不存在返回 nil。
func (p *Palace) FindStar(name string) *Star {
	for i := range p.Stars {
		if p.Stars[i].Name == name {
			return &p.Stars[i]
		}
	}
	return nil
}

// MajorStarNames 宫内主星名列表。
func (p *Palace) MajorStarNames() []string {
	var out []string
	for i := range p.Stars {
		if p.Stars[i].Type == StarMajor {
			out = append(out, p.Stars[i].Name)
		}
	}
	return out
}
