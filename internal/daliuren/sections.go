// 大六壬分节深断层:课体详解/三传始末/天将所临/应期推算(确定性,不走 LLM)。
// 断语依《六壬大全》《课经》通行义理归纳、原创行文;与 208 例断案回归的
// 起课口径同源,只做呈现纵深,不改任何推演。
package daliuren

import (
	"fmt"
	"strings"
)

// JudgeSection 断语分节(免费确定性层的呈现单元)。
type JudgeSection struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

// keTypeDeep 九门课体详解:古义 + 事象 + 行动之方(义引归纳,原创行文)。
var keTypeDeep = map[string]string{
	"贼克": "贼克为九课之首:四课之中上下有克,下贼上谓之「始入」,上克下谓之「元首」。有克则有事,吉凶分明而来意直——问事者心中必有一桩明确的争执、决断或攻守。此课应期最速,事势已成短兵相接,拖延无益;当机立断、以刚制之,先动者得势。",
	"比用": "比用课:四课之中克者不止一处,难以皆用,故取与日干俱比(阴阳同类)者发用。事象是多头并起、须择一路而行——同侪之助、亲熟之门是本课的解法。谋事勿走生僻门径,循类相求、托同气之人,和同则济;若所比之人自身失位,则改换门庭再图。",
	"涉害": "涉害课:诸课皆克而俱比或俱不比,取上神涉历地盘深浅而定发用。涉浅者事易而应速,涉深者得力迟而根基厚。此课之事必经磨折,过程曲折乃是命题本身——非阻在人,而阻在路径必须绕行。宜存长算、分段推进,每过一节气则势进一分,耐历炼方见归着。",
	"遥克": "遥克课:四课上下皆无克,而第二、三、四课之神与日干遥相克应——先取神克干者(蒿矢),次取干克神者(弹射)。事自远方、外界、间接而来,如隔空发矢,声势大而力道轻。多主虚惊、传闻、远方之扰;察明来路,十之七八可安,不必闻风而自乱阵脚。",
	"昴星": "昴星课:四课无克无遥,取昴星法——阳日取酉宫上神仰视,阴日取酉宫地盘俯察。酉为昴宿之位,虎视之象,事出非常、多起于所不见之处。主意外、惊扰、暗中窥伺;此课最忌妄动,宜静守本位、明哨暗防,待事象自现再作应对。",
	"别责": "别责课:三课备(四课有重复)而无克无遥,正路不通,别取一神为用——刚日取干合之上神,柔日取支前三合。事有缺憾不全,如器缺一角;正面强攻无门,须借他力、走旁径、托中介,曲折乃成。成事之后亦须补其缺,否则遗患。",
	"八专": "八专课:干支同位(甲寅、庚申、丁未、己未、癸丑五日),干支同宫、上下一体。公私相混、二人同心之象——同心则其利断金,合伙、共谋、夫妇同事皆宜;然上下不分亦是本课之弊,防狎昵不明、权责不清、内外无别之失,先立规矩再共事。",
	"伏吟": "伏吟课:天地盘完全重合,天神各伏本宫。伏者不动,事主静止、忧疑滞塞——外象无事而内里暗结。谋为难展,强动则自损;守旧待时是正解。其动机在「冲」:待冲开伏局之日(与用神相冲之期),滞者自通。",
	"返吟": "返吟课:天地盘正对冲,神皆居冲位。返复之象——事去而复来、成而复散、人往而复返。动荡不安,中途生变是常态而非例外;凡约期、行旅、交易皆宜留退路。一动不如一静,若不得不动,则以短程、分段、可回头者为宜。",
}

// jiangDeep 十二天将深断:将义 + 所主事类(通行壬学类神义,原创行文)。
var jiangDeep = map[string]string{
	"贵人": "十二将之主,土将而尊。主贵人提携、尊长关照、官贵之力;临传则凶中有解、事得高位者转圜。昼贵得力于明面,夜贵得力于私下。",
	"螣蛇": "火将,主惊恐怪异、虚惊缠绕、梦寐不宁。临传多有无形之扰——传言、疑心、怪事;实祸常小于其声势,安神定志则蛇化为绳。",
	"朱雀": "火将,主文书、信息、口舌。吉则文章科名、音信文契之喜;凶则口舌是非、官讼文书之累。临传所问之事必与「言语文字」相关。",
	"六合": "木将,主和合、交易、婚媾、中人。临传事有撮合之力、得中间人之助;和合之中亦防苟合,契约往来须明。",
	"勾陈": "土将,主勾连、迟滞、田土、争讼。临传事被勾缠难解、进度拖延;涉田宅地界之争尤验。解在厘清头绪、各归其位。",
	"青龙": "木将,吉将之最。主财帛、喜庆、生气。临传财喜之应显然,谋望可进;龙贵生旺之地其力尤宏。",
	"天空": "土将,主虚诈、空言、无实。临传所遇之人事多虚少实——承诺难兑、文书不实;凡事验其实据,勿听空言。",
	"白虎": "金将,主疾病、伤灾、道路、丧服。临传最须谨慎之将——主血光刀兵、疾厄凶信;亦主道路奔驰,出行者尤当慎。",
	"太常": "土将,主宴饮、衣冠、文物、礼聘。临传事关礼尚往来、饮食衣冠;吉则受赏得聘,平则酬酢应付而已。",
	"玄武": "水将,主盗贼、阴私、遗失、欺瞒。临传防暗中之失——失物、被窃、被瞒;亦主阴私暧昧之事,防人亦防己涉。",
	"太阴": "金将,主阴私庇护、妇人、暗助。临传得暗中之助、阴人之力;事宜密行,不宜张扬,阴谋阴成。",
	"天后": "水将,主妇人、恩泽、阴柔之力。临传事关女性、内宅、恩私;柔道可成,刚求反拙。",
}

// liuHe 地支六合对(子丑/寅亥/卯戌/辰酉/巳申/午未)。
var liuHe = map[int]int{0: 1, 1: 0, 2: 11, 11: 2, 3: 10, 10: 3, 4: 9, 9: 4, 5: 8, 8: 5, 6: 7, 7: 6}

// buildSections 组装分节深断(在 Judge 推演完吉凶后调用)。
func (r *Result) buildSections(j *Judgment) {
	// ── 课体详解 ──
	var kb strings.Builder
	deep := keTypeDeep[r.KeType]
	if deep == "" {
		deep = j.KeTypeText
	}
	kb.WriteString(deep)
	gui := "夜贵"
	if r.GuiIsDay {
		gui = "昼贵"
	}
	kb.WriteString(fmt.Sprintf("本课以%s%s日%s时起,月将%s加时布天地盘,用%s;旬空%s%s。",
		r.DayStem, r.DayBranch, r.HourBranch, r.MonthGen, gui, r.XunKong[0], r.XunKong[1]))
	j.Sections = append(j.Sections, JudgeSection{Key: "keti", Title: "课体详解", Text: kb.String()})

	// ── 三传始末:逐传叙事(支性+乘将深断+遁干/空亡) ──
	frames := []string{"初传为发端,事之来由", "中传为经过,事之转折", "末传为归宿,事之结局"}
	var sc strings.Builder
	sc.WriteString("三传者,事之始末也。")
	for i := 0; i < 3 && i < len(r.Chuan); i++ {
		ci := branchIdx(r.Chuan[i])
		if ci < 0 {
			continue
		}
		sc.WriteString(fmt.Sprintf("%s——%s(%s%s)",
			frames[i], r.Chuan[i], yinYangName(ci), elementNames[branchElement[ci]]))
		if r.ChuanDunGan[i] != "" {
			sc.WriteString(",遁干" + r.ChuanDunGan[i])
		} else if r.XunKong[0] != "" {
			sc.WriteString(",落旬空(传空事虚,出空填实方见分晓)")
		}
		if jg := r.ChuanJiang[i]; jg != "" {
			sc.WriteString(fmt.Sprintf(",乘%s。", jg))
		} else {
			sc.WriteString("。")
		}
	}
	j.Sections = append(j.Sections, JudgeSection{Key: "sanchuan", Title: "三传始末", Text: sc.String()})

	// ── 天将所临:三传之将逐一深断(凡壬课吉凶系于天将) ──
	var tj strings.Builder
	tj.WriteString("凡壬课吉凶,系于天将。")
	seen := map[string]bool{}
	names := []string{"初传", "中传", "末传"}
	for i := 0; i < 3 && i < len(r.ChuanJiang); i++ {
		jg := r.ChuanJiang[i]
		if jg == "" || seen[jg] {
			continue
		}
		seen[jg] = true
		if d := jiangDeep[jg]; d != "" {
			tj.WriteString(fmt.Sprintf("%s乘【%s】:%s", names[i], jg, d))
		}
	}
	j.Sections = append(j.Sections, JudgeSection{Key: "tianjiang", Title: "天将所临", Text: tj.String()})

	// ── 应期推算:初传冲合之日 + 空亡出空 ──
	var yq strings.Builder
	if ci := branchIdx(r.Chuan[0]); ci >= 0 {
		chong := branches[(ci+6)%12]
		he := branches[liuHe[ci]]
		yq.WriteString(fmt.Sprintf("应期以初传%s为候:急应看冲——%c日(%s之冲)动机一到、事象即现;缓应看合——%c日(%s之合)缔结落定。",
			r.Chuan[0], chong, r.Chuan[0], he, r.Chuan[0]))
	}
	kongChuan := false
	for i := 0; i < 3; i++ {
		if r.ChuanDunGan[i] == "" && r.XunKong[0] != "" {
			kongChuan = true
		}
	}
	if kongChuan {
		yq.WriteString(fmt.Sprintf("传中带空(旬空%s%s):空亡之事以「出空」为期,过此一旬、或值空之支填实之日,虚者转实。",
			r.XunKong[0], r.XunKong[1]))
	}
	switch r.KeType {
	case "伏吟":
		yq.WriteString("伏吟之动机专在冲开:与用神相冲之日时,滞局自解。")
	case "返吟":
		yq.WriteString("返吟往复,应期常见两度:初动一验、复动再验,首尾之期皆当留意。")
	}
	j.Sections = append(j.Sections, JudgeSection{Key: "yingqi", Title: "应期推算", Text: yq.String()})
}

// yinYangName 支阴阳。
func yinYangName(branchIdx int) string {
	if branchIdx%2 == 0 {
		return "阳"
	}
	return "阴"
}
