// Package knowledge 倪海厦三纪知识库 + 紫微业务知识(主星描述、合盘断语、
// 城市经度、名人盘)。数据源为 data.FS 内嵌 JSON,启动加载后只读,并发安全。
package knowledge

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

// Base 知识库总入口。加载后全部字段只读。
type Base struct {
	// 三纪原始 JSON(结构化透传给 API)
	Tianji json.RawMessage
	Renji  json.RawMessage
	Diji   json.RawMessage
	Bio    json.RawMessage

	// 已解析的常用视图
	TianjiQuotes []Quote // 倪师语录
	StarDesc     map[string]StarDescription
	StarSlugs    map[string]string // 主星名 → 拼音 slug
	StarOrder    []string          // 十四主星顺序
	// StarLore 全量星曜档案(主星/辅煞/杂曜,义引归纳原创行文);
	// StarCycles 四大十二神逐名义(changsheng12/boshi12/suiqian12/jiangqian12);
	// StarFlow 流曜十义(按去前缀后的字键:魁钺昌曲禄羊陀马鸾喜)。
	StarLore   map[string]StarLore
	StarCycles map[string]map[string]string
	StarFlow   map[string]string
	Topics     TopicMeta
	Heming     HemingKnowledge
	Provinces  []Province
	WorldCities []WorldCity
	Famous      []FamousPerson
}

// Quote 倪师语录。
type Quote struct {
	Text  string `json:"text"`
	Topic string `json:"topic,omitempty"`
}

// StarDescription 主星速览(倪海厦体系)。
type StarDescription struct {
	Keywords string `json:"keywords"`
	Nature   string `json:"nature"`
	Element  string `json:"element"`
}

// StarLore 星曜档案:五行/化气/主司为传统口径,义理为义引归纳原创行文。
// 杂曜多无五行化气,仅 si+gist。
type StarLore struct {
	Element string `json:"element,omitempty"` // 五行(如 己土)
	Hua     string `json:"hua,omitempty"`     // 化气(如 化气曰尊)
	Si      string `json:"si"`                // 主司(如 官禄主 · 帝座)
	Gist    string `json:"gist"`              // 义理档案
}

// TopicMeta 解读主题标签。
type TopicMeta struct {
	PalaceName map[string]string `json:"palaceName"` // topic → 宫位名
	Label      map[string]string `json:"label"`      // topic → 中文标签
	Order      []string          `json:"order"`
}

// HemingKnowledge 合盘知识库。
type HemingKnowledge struct {
	StarInFuqi         map[string]StarInFuqi `json:"starInFuqi"`
	SihuaInFuqi        map[string]string     `json:"sihuaInFuqi"`
	Methodology        string                `json:"methodology"`
	MarriageStarsBrief map[string]string     `json:"marriageStarsBrief"`
	ScoreCriteria      json.RawMessage       `json:"scoreCriteria"`
}

// StarInFuqi 十四主星在夫妻宫断语。
type StarInFuqi struct {
	Summary      string `json:"summary"`
	Good         string `json:"good"`
	Bad          string `json:"bad"`
	SpouseTraits string `json:"spouse_traits"`
	Timing       string `json:"timing"`
	NiQuote      string `json:"ni_quote,omitempty"`
}

// Province 省份与城市经度(真太阳时校正)。
type Province struct {
	Name   string `json:"name"`
	Cities []City `json:"cities"`
}

// City 城市经度。
type City struct {
	Name      string  `json:"name"`
	Longitude float64 `json:"longitude"`
}

// WorldCity 世界主要城市:经度 + 标准时区 UTC 偏移(小时),用于国际真太阳时。
type WorldCity struct {
	Country   string  `json:"country"`
	Name      string  `json:"name"`
	Longitude float64 `json:"longitude"`
	UTCOffset float64 `json:"utcOffset"`
	DST       bool    `json:"dst,omitempty"` // 该地实行夏令时,夏令时出生须另减 1 小时
}

// FamousPerson 名人命盘样例。
type FamousPerson struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Year        int    `json:"year"`
	Month       int    `json:"month"`
	Day         int    `json:"day"`
	Hour        int    `json:"hour"`
	Gender      string `json:"gender"`
	Notable     string `json:"notable"`
}

// Load 从内嵌数据文件系统加载知识库。
func Load(dataFS fs.FS) (*Base, error) {
	b := &Base{}

	read := func(path string) (json.RawMessage, error) {
		raw, err := fs.ReadFile(dataFS, path)
		if err != nil {
			return nil, fmt.Errorf("读取 %s 失败: %w", path, err)
		}
		return raw, nil
	}

	var err error
	if b.Tianji, err = read("nihai/tianji.json"); err != nil {
		return nil, err
	}
	if b.Renji, err = read("nihai/renji.json"); err != nil {
		return nil, err
	}
	if b.Diji, err = read("nihai/diji.json"); err != nil {
		return nil, err
	}
	if b.Bio, err = read("nihai/bio.json"); err != nil {
		return nil, err
	}

	var tianjiView struct {
		Quotes []Quote `json:"quotes"`
	}
	if err := json.Unmarshal(b.Tianji, &tianjiView); err != nil {
		return nil, fmt.Errorf("解析天纪语录失败: %w", err)
	}
	b.TianjiQuotes = tianjiView.Quotes

	starsRaw, err := read("knowledge/stars.json")
	if err != nil {
		return nil, err
	}
	var starsView struct {
		Descriptions map[string]StarDescription   `json:"descriptions"`
		Slugs        map[string]string            `json:"slugs"`
		Order        []string                     `json:"order"`
		Lore         map[string]StarLore          `json:"lore"`
		Cycles       map[string]map[string]string `json:"cycles"`
		Flow         map[string]string            `json:"flow"`
	}
	if err := json.Unmarshal(starsRaw, &starsView); err != nil {
		return nil, fmt.Errorf("解析主星知识失败: %w", err)
	}
	b.StarDesc = starsView.Descriptions
	b.StarSlugs = starsView.Slugs
	b.StarOrder = starsView.Order
	b.StarLore = starsView.Lore
	b.StarCycles = starsView.Cycles
	b.StarFlow = starsView.Flow

	topicsRaw, err := read("knowledge/topics.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(topicsRaw, &b.Topics); err != nil {
		return nil, fmt.Errorf("解析主题标签失败: %w", err)
	}

	hemingRaw, err := read("knowledge/heming.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(hemingRaw, &b.Heming); err != nil {
		return nil, fmt.Errorf("解析合盘知识失败: %w", err)
	}

	citiesRaw, err := read("cities.json")
	if err != nil {
		return nil, err
	}
	var citiesView struct {
		Provinces []Province `json:"provinces"`
	}
	if err := json.Unmarshal(citiesRaw, &citiesView); err != nil {
		return nil, fmt.Errorf("解析城市数据失败: %w", err)
	}
	b.Provinces = citiesView.Provinces

	worldRaw, err := read("world_cities.json")
	if err != nil {
		return nil, err
	}
	var worldView struct {
		Cities []WorldCity `json:"cities"`
	}
	if err := json.Unmarshal(worldRaw, &worldView); err != nil {
		return nil, fmt.Errorf("解析世界城市数据失败: %w", err)
	}
	b.WorldCities = worldView.Cities

	famousRaw, err := read("famous.json")
	if err != nil {
		return nil, err
	}
	var famousView struct {
		Persons []FamousPerson `json:"persons"`
	}
	if err := json.Unmarshal(famousRaw, &famousView); err != nil {
		return nil, fmt.Errorf("解析名人数据失败: %w", err)
	}
	b.Famous = famousView.Persons

	return b, nil
}

// LongitudeOf 按省市名查经度,未收录返回 0。
func (b *Base) LongitudeOf(province, city string) float64 {
	for _, p := range b.Provinces {
		if p.Name != province {
			continue
		}
		for _, c := range p.Cities {
			if c.Name == city {
				return c.Longitude
			}
		}
	}
	return 0
}

// WorldCityOf 按城市名查世界城市(经度 + UTC 偏移),未收录返回 nil。
func (b *Base) WorldCityOf(name string) *WorldCity {
	for i := range b.WorldCities {
		if b.WorldCities[i].Name == name {
			return &b.WorldCities[i]
		}
	}
	return nil
}
