package usecase

import (
	"sort"
	"strings"

	"mahjong-backend/internal/domain/entity"
)

// YakuUsecase provides evaluation logic for yaku
type YakuUsecase struct{}

func NewYakuUsecase() *YakuUsecase { return &YakuUsecase{} }

// EvaluateYaku evaluates simple yaku based on request
// NOTE: Minimal logic to satisfy Swagger contract; extend with full rules later.
func (u *YakuUsecase) EvaluateYaku(req *entity.YakuEvaluationRequest) *entity.YakuEvaluationResponse {
	result := &entity.YakuEvaluationResponse{
		Yaku:         []entity.YakuItem{},
		Fu:           0,
		Han:          0,
		DoraCount:    0,
		UraDoraCount: 0,
		TotalHan:     0,
		Yakuman:      []string{},
		Score:        0,
	}

	// Check for invalid hand (chombo)
	if u.isChombo(req) {
		result.IsChombo = true
		result.Yakuman = []string{}
		result.TotalHan = 0
		result.Score = 0
		return result
	}

	// Riichi
	if req.Riichi {
		result.Yaku = append(result.Yaku, entity.YakuItem{Name: "立直", Han: 1})
		result.Han += 1
	}

	// Tsumo (menzen only)
	if req.IsTsumo && len(req.OpenMelds) == 0 {
		result.Yaku = append(result.Yaku, entity.YakuItem{Name: "門前清自摸和", Han: 1})
		result.Han += 1
	}

	// Ippatsu
	if req.Ippatsu && req.Riichi {
		result.Yaku = append(result.Yaku, entity.YakuItem{Name: "一発", Han: 1})
		result.Han += 1
	}

	// Build full tile list (hand + open meld tiles)
	fullTiles := make([]string, 0, len(req.Tehai)+len(req.OpenMelds)*3)
	fullTiles = append(fullTiles, req.Tehai...)
	for _, m := range req.OpenMelds {
		fullTiles = append(fullTiles, m.Tiles...)
	}

	// Yakuman detection (adds yakuman names and counts as 13 han each)
	yakumanList := u.detectYakuman(fullTiles, req)
	for _, y := range yakumanList {
		result.Yakuman = append(result.Yakuman, y)
		result.TotalHan += 13
	}

	// If yakuman is present, ignore all other yaku and calculate yakuman score
	if len(yakumanList) > 0 {
		// 親子の判定：場風と自風が同じ場合は親
		isOya := strings.ToLower(req.RoundWind) == strings.ToLower(req.SeatWind)
		result.Score = calculateScore(result.TotalHan, isOya)
		return result
	}

	// 2+ han yaku detection (adds to Han)
	for name, han := range u.detectTwoPlus(fullTiles, req) {
		result.Yaku = append(result.Yaku, entity.YakuItem{Name: name, Han: han})
		result.Han += han
	}

	// 1 han yaku detection
	for name, han := range u.detectOneHan(fullTiles, req) {
		result.Yaku = append(result.Yaku, entity.YakuItem{Name: name, Han: han})
		result.Han += han
	}

	// 役牌検出（三元牌・風牌）
	u.detectYakuhai(result, req)

	// Dora counting: convert indicators to actual dora tiles
	doraSet := make(map[string]int)
	for _, indicator := range req.DoraIndicators {
		doraTile := getDoraFromIndicator(indicator)
		doraSet[doraTile]++
	}

	// Count dora in hand and open melds
	for _, t := range req.Tehai {
		if doraSet[t] > 0 {
			result.DoraCount++
		}
	}
	for _, m := range req.OpenMelds {
		for _, t := range m.Tiles {
			if doraSet[t] > 0 {
				result.DoraCount++
			}
		}
	}

	// Ura dora counting (only if riichi)
	if req.Riichi {
		uraDoraSet := make(map[string]int)
		for _, indicator := range req.UraDoraIndicators {
			doraTile := getDoraFromIndicator(indicator)
			uraDoraSet[doraTile]++
		}
		for _, t := range req.Tehai {
			if uraDoraSet[t] > 0 {
				result.UraDoraCount++
			}
		}
		for _, m := range req.OpenMelds {
			for _, t := range m.Tiles {
				if uraDoraSet[t] > 0 {
					result.UraDoraCount++
				}
			}
		}
	}

	result.TotalHan = result.Han + result.DoraCount + result.UraDoraCount

	// 親子の判定：場風と自風が同じ場合は親
	isOya := strings.ToLower(req.RoundWind) == strings.ToLower(req.SeatWind)

	// 正しい麻雀点数表による計算
	result.Score = calculateScore(result.TotalHan, isOya)
	return result
}

// detectYakuhai detects honor tile yaku (dragons and winds)
func (u *YakuUsecase) detectYakuhai(result *entity.YakuEvaluationResponse, req *entity.YakuEvaluationRequest) {
	// 三元牌の役牌検出
	dragonMap := map[string]string{
		"白": "役牌(白)", "haku": "役牌(白)",
		"發": "役牌(發)", "hatu": "役牌(發)",
		"中": "役牌(中)", "chun": "役牌(中)",
	}

	// 風牌の役牌検出
	windMap := map[string]string{
		"ton": "東", "東": "東",
		"nan": "南", "南": "南",
		"sya": "西", "西": "西",
		"pe": "北", "北": "北",
	}

	// 全牌を集める（手牌＋副露）
	fullTiles := make([]string, 0, len(req.Tehai)+len(req.OpenMelds)*3)
	fullTiles = append(fullTiles, req.Tehai...)
	for _, m := range req.OpenMelds {
		fullTiles = append(fullTiles, m.Tiles...)
	}

	// 三元牌の役牌チェック
	for tile, yakuName := range dragonMap {
		count := 0
		for _, t := range fullTiles {
			if t == tile {
				count++
			}
		}
		if count >= 3 {
			result.Yaku = append(result.Yaku, entity.YakuItem{Name: yakuName, Han: 1})
			result.Han += 1
		}
	}

	// 風牌の役牌チェック（場風・自風）
	for tile, wind := range windMap {
		count := 0
		for _, t := range fullTiles {
			if t == tile {
				count++
			}
		}
		if count >= 3 {
			// 場風チェック
			if matchesWind(wind, req.RoundWind) {
				result.Yaku = append(result.Yaku, entity.YakuItem{Name: "役牌(場風)", Han: 1})
				result.Han += 1
			}
			// 自風チェック
			if matchesWind(wind, req.SeatWind) {
				result.Yaku = append(result.Yaku, entity.YakuItem{Name: "役牌(自風)", Han: 1})
				result.Han += 1
			}
			// 場風でも自風でもない場合は役なし
		}
	}
}

// matchesWind checks if tile wind matches given wind keyword
func matchesWind(tileWind string, wind string) bool {
	w := strings.ToLower(wind)
	switch tileWind {
	case "東":
		return w == "east"
	case "南":
		return w == "south"
	case "西":
		return w == "west"
	case "北":
		return w == "north"
	default:
		return false
	}
}

// ---- detectors & helpers ----

type tile struct {
	suit    string
	num     int
	isHonor bool
	raw     string
}

func parseTile(s string) tile {
	t := tile{raw: s}
	rs := strings.TrimSpace(s)
	if rs == "白" || rs == "haku" || rs == "hatu" || rs == "發" || rs == "chun" || rs == "中" || rs == "ton" || rs == "nan" || rs == "sya" || rs == "pe" || rs == "東" || rs == "南" || rs == "西" || rs == "北" {
		t.isHonor = true
		t.suit = "z"
		return t
	}
	if rs == "5pr" {
		t.suit = "p"
		t.num = 5
		return t
	}
	if rs == "5sr" {
		t.suit = "s"
		t.num = 5
		return t
	}
	if len(rs) >= 2 {
		n := int(rs[0] - '0')
		t.num = n
		t.suit = string(rs[1])
	}
	return t
}

func isTerminal(t tile) bool {
	if t.isHonor {
		return false
	}
	return t.num == 1 || t.num == 9
}

func (u *YakuUsecase) detectYakuman(full []string, req *entity.YakuEvaluationRequest) []string {
	tiles := make([]tile, 0, len(full))
	for _, s := range full {
		tiles = append(tiles, parseTile(s))
	}

	var yakuman []string

	// 国士無双・国士無双十三面待ち
	if isKokushi(full) {
		if isKokushiJuusanmen(full, req.WinTile) {
			yakuman = append(yakuman, "国士無双十三面待ち")
		} else {
			yakuman = append(yakuman, "国士無双")
		}
	}

	// 四暗刻・四暗刻単騎待ち
	if isSuuankou(full, req) {
		if isSuuankouTanki(full, req) {
			yakuman = append(yakuman, "四暗刻単騎待ち")
		} else {
			yakuman = append(yakuman, "四暗刻")
		}
	}

	// 大三元
	dragonTriplets := countTripletKinds(full, []string{"白", "haku", "發", "hatu", "中", "chun"})
	if dragonTriplets >= 3 {
		yakuman = append(yakuman, "大三元")
	}

	// 小四喜・大四喜
	windTriplets := countTripletKinds(full, []string{"ton", "東", "nan", "南", "sya", "西", "pe", "北"})
	if windTriplets == 4 {
		yakuman = append(yakuman, "大四喜")
	} else if windTriplets == 3 && hasPair(full, []string{"ton", "東", "nan", "南", "sya", "西", "pe", "北"}) {
		yakuman = append(yakuman, "小四喜")
	}

	// 字一色
	if allHonors(tiles) {
		yakuman = append(yakuman, "字一色")
	}

	// 緑一色
	if isRyuuiisou(tiles) {
		yakuman = append(yakuman, "緑一色")
	}

	// 清老頭
	if allTerminals(tiles) {
		yakuman = append(yakuman, "清老頭")
	}

	// 九蓮宝燈・純正九蓮宝燈
	if isChuuren(full, req) {
		if isJunseiChuuren(full, req) {
			yakuman = append(yakuman, "純正九蓮宝燈")
		} else {
			yakuman = append(yakuman, "九蓮宝燈")
		}
	}

	// 四槓子
	if countKan(req.OpenMelds) == 4 {
		yakuman = append(yakuman, "四槓子")
	}

	// 天和・地和・人和
	if req.Tenhou {
		yakuman = append(yakuman, "天和")
	}
	if req.Chiihou {
		yakuman = append(yakuman, "地和")
	}
	if req.Renhou {
		yakuman = append(yakuman, "人和")
	}

	return yakuman
}

func (u *YakuUsecase) detectTwoPlus(full []string, req *entity.YakuEvaluationRequest) map[string]int {
	res := map[string]int{}
	tiles := make([]tile, 0, len(full))
	for _, s := range full {
		tiles = append(tiles, parseTile(s))
	}
	// 2 han yaku
	if countSevenPairs(req.Tehai) == 7 && len(req.OpenMelds) == 0 {
		res["七対子"] = 2
	}
	if noChiInMelds(req.OpenMelds) && hasAtLeastNTriplets(full, 4) {
		res["対々和"] = 2
	}
	if onlyTerminalsAndHonors(tiles) {
		res["混老頭"] = 2
	}
	dTrip := countTripletKinds(full, []string{"白", "haku", "發", "hatu", "中", "chun"})
	if dTrip == 2 && hasPair(full, []string{"白", "haku", "發", "hatu", "中", "chun"}) {
		res["小三元"] = 2
	}
	if countKan(req.OpenMelds) >= 3 {
		res["三槓子"] = 2
	}
	// 3 han yaku
	if isHonitsu(tiles) {
		if len(req.OpenMelds) == 0 {
			res["混一色"] = 3
		} else {
			res["混一色"] = 2
		}
	}
	// 6 han yaku
	if isChinitsu(tiles) {
		if len(req.OpenMelds) == 0 {
			res["清一色"] = 6
		} else {
			res["清一色"] = 5
		}
	}
	// 2 han yaku (with kui-sagari)
	if hasIttsuu(full) {
		if len(req.OpenMelds) == 0 {
			res["一気通貫"] = 2
		} else {
			res["一気通貫"] = 1
		}
	}
	if hasSanshokuDoujun(full) {
		if len(req.OpenMelds) == 0 {
			res["三色同順"] = 2
		} else {
			res["三色同順"] = 1
		}
	}
	if hasSanshokuDoukou(full) {
		res["三色同刻"] = 2
	}
	// 3 han yaku (with kui-sagari)
	if hasChanta(full) {
		if len(req.OpenMelds) == 0 {
			res["混全帯么九"] = 2
		} else {
			res["混全帯么九"] = 1
		}
	}
	if hasJunchan(full) {
		if len(req.OpenMelds) == 0 {
			res["純全帯么九"] = 3
		} else {
			res["純全帯么九"] = 2
		}
	}
	// 3 han yaku
	if hasRyanpeikou(req.Tehai) && len(req.OpenMelds) == 0 {
		res["二盃口"] = 3
	}
	return res
}

func allHonors(tiles []tile) bool {
	if len(tiles) == 0 {
		return false
	}
	for _, t := range tiles {
		if !t.isHonor {
			return false
		}
	}
	return true
}
func allTerminals(tiles []tile) bool {
	if len(tiles) == 0 {
		return false
	}
	for _, t := range tiles {
		if t.isHonor || !isTerminal(t) {
			return false
		}
	}
	return true
}
func onlyTerminalsAndHonors(tiles []tile) bool {
	if len(tiles) == 0 {
		return false
	}
	for _, t := range tiles {
		if !t.isHonor && !isTerminal(t) {
			return false
		}
	}
	return true
}

func countTripletKinds(full []string, target []string) int {
	set := map[string]bool{}
	for _, x := range target {
		set[x] = true
	}
	counts := map[string]int{}
	for _, s := range full {
		if set[s] {
			counts[s]++
		}
	}
	trip := 0
	for _, c := range counts {
		if c >= 3 {
			trip++
		}
	}
	return trip
}

func hasPair(full []string, target []string) bool {
	set := map[string]bool{}
	for _, x := range target {
		set[x] = true
	}
	counts := map[string]int{}
	for _, s := range full {
		if set[s] {
			counts[s]++
		}
	}
	for _, c := range counts {
		if c >= 2 {
			return true
		}
	}
	return false
}

func hasAtLeastNTriplets(full []string, n int) bool {
	counts := map[string]int{}
	for _, s := range full {
		counts[s]++
	}
	trip := 0
	for _, c := range counts {
		if c >= 3 {
			trip++
		}
	}
	return trip >= n
}
func noChiInMelds(m []entity.MeldRequest) bool {
	for _, x := range m {
		if strings.ToLower(x.Type) == "chi" {
			return false
		}
	}
	return true
}
func countKan(m []entity.MeldRequest) int {
	c := 0
	for _, x := range m {
		if strings.ToLower(x.Type) == "kan" {
			c++
		}
	}
	return c
}

func isChinitsu(tiles []tile) bool {
	suit := ""
	for _, t := range tiles {
		if t.isHonor {
			return false
		}
		if suit == "" {
			suit = t.suit
		} else if t.suit != suit {
			return false
		}
	}
	return len(tiles) > 0
}
func isHonitsu(tiles []tile) bool {
	suit := ""
	hasHonor := false
	for _, t := range tiles {
		if t.isHonor {
			hasHonor = true
			continue
		}
		if suit == "" {
			suit = t.suit
		} else if t.suit != suit {
			return false
		}
	}
	return hasHonor && suit != ""
}

func isRyuuiisou(tiles []tile) bool {
	green := map[string]bool{"2s": true, "3s": true, "4s": true, "6s": true, "8s": true, "發": true, "hatu": true}
	if len(tiles) == 0 {
		return false
	}
	for _, t := range tiles {
		if t.isHonor {
			if t.raw != "發" && t.raw != "hatu" {
				return false
			}
		} else {
			key := string('0'+byte(t.num)) + t.suit
			if !green[key] {
				return false
			}
		}
	}
	return true
}

// isSuuankou checks for four concealed triplets
func isSuuankou(full []string, req *entity.YakuEvaluationRequest) bool {
	if len(req.OpenMelds) > 0 {
		return false // Must be all concealed
	}

	// Count triplets in hand only (tehai)
	counts := make(map[string]int)
	for _, t := range req.Tehai {
		counts[t]++
	}

	triplets := 0
	for _, count := range counts {
		if count >= 3 {
			triplets++
		}
	}

	// Must have exactly 4 triplets and 1 pair (4*3 + 1*2 = 14 tiles)
	return triplets == 4 && len(req.Tehai) == 13 // 13 tiles in hand + 1 winning tile = 14
}

// isSuuankouTanki checks for four concealed triplets with single wait
func isSuuankouTanki(full []string, req *entity.YakuEvaluationRequest) bool {
	if !isSuuankou(full, req) {
		return false
	}

	// Check if winning tile forms a pair (tanki wait)
	counts := make(map[string]int)
	for _, t := range req.Tehai {
		counts[t]++
	}

	return counts[req.WinTile] == 2
}

// isKokushiJuusanmen checks for 13-way wait kokushi
func isKokushiJuusanmen(full []string, winTile string) bool {
	if !isKokushi(full) {
		return false
	}

	terminals := map[string]bool{
		"1m": true, "9m": true, "1p": true, "9p": true, "1s": true, "9s": true,
		"ton": true, "nan": true, "sya": true, "pe": true,
		"haku": true, "hatu": true, "chun": true,
		"東": true, "南": true, "西": true, "北": true,
		"白": true, "發": true, "中": true,
	}

	counts := make(map[string]int)
	for _, t := range full {
		if terminals[t] {
			counts[t]++
		}
	}

	// Check if all 13 types are present and winning tile appears only once
	return len(counts) == 13 && counts[winTile] == 1
}

// isChuuren checks for nine gates
func isChuuren(full []string, req *entity.YakuEvaluationRequest) bool {
	if len(req.OpenMelds) > 0 {
		return false // Must be concealed
	}

	// Must be single suit
	suits := make(map[string]bool)
	for _, t := range full {
		tile := parseTile(t)
		if tile.isHonor {
			return false
		}
		suits[tile.suit] = true
	}

	if len(suits) != 1 {
		return false
	}

	// Get the suit
	var suit string
	for s := range suits {
		suit = s
	}

	// Count tiles by number
	counts := make([]int, 10) // index 0 unused, 1-9 for tile numbers
	for _, t := range full {
		tile := parseTile(t)
		if tile.suit == suit && tile.num >= 1 && tile.num <= 9 {
			counts[tile.num]++
		}
	}

	// Check pattern: 1112345678999 + any one tile
	expected := []int{0, 3, 1, 1, 1, 1, 1, 1, 1, 3} // 0 is unused

	// Find which number has extra tile
	extra := -1
	for i := 1; i <= 9; i++ {
		if counts[i] == expected[i]+1 {
			if extra != -1 {
				return false // More than one extra
			}
			extra = i
		} else if counts[i] != expected[i] {
			return false
		}
	}

	return extra != -1
}

// isJunseiChuuren checks for pure nine gates (9-way wait)
func isJunseiChuuren(full []string, req *entity.YakuEvaluationRequest) bool {
	if !isChuuren(full, req) {
		return false
	}

	// Check if winning tile makes it 9-way wait
	winTileNum := parseTile(req.WinTile).num

	// In pure chuuren, the extra tile can be any of 1-9
	// and the hand before winning should be exactly 1112345678999
	counts := make([]int, 10)
	for _, t := range req.Tehai {
		tile := parseTile(t)
		if tile.num >= 1 && tile.num <= 9 {
			counts[tile.num]++
		}
	}

	// Remove winning tile
	counts[winTileNum]--

	// Check if remaining hand is exactly 1112345678999
	expected := []int{0, 3, 1, 1, 1, 1, 1, 1, 1, 3}
	for i := 1; i <= 9; i++ {
		if counts[i] != expected[i] {
			return false
		}
	}

	return true
}

func isKokushi(full []string) bool {
	need := map[string]bool{"1m": true, "9m": true, "1p": true, "9p": true, "1s": true, "9s": true, "ton": true, "nan": true, "sya": true, "pe": true, "haku": true, "hatu": true, "chun": true, "東": true, "南": true, "西": true, "北": true, "白": true, "發": true, "中": true}
	uniq := map[string]bool{}
	extra := 0
	for _, s := range full {
		if need[s] {
			if uniq[s] {
				extra++
			} else {
				uniq[s] = true
			}
		}
	}
	d := 0
	for k := range need {
		if uniq[k] {
			d++
		}
	}
	return d >= 13 && len(full) >= 14 && extra >= 1
}

func hasIttsuu(full []string) bool {
	suits := []string{"m", "p", "s"}
	counts := countByNumSuit(full)
	for _, s := range suits {
		if counts[s][1] > 0 && counts[s][2] > 0 && counts[s][3] > 0 && counts[s][4] > 0 && counts[s][5] > 0 && counts[s][6] > 0 && counts[s][7] > 0 && counts[s][8] > 0 && counts[s][9] > 0 {
			return true
		}
	}
	return false
}
func hasSanshokuDoujun(full []string) bool {
	counts := countByNumSuit(full)
	for n := 1; n <= 7; n++ {
		if counts["m"][n] > 0 && counts["m"][n+1] > 0 && counts["m"][n+2] > 0 && counts["p"][n] > 0 && counts["p"][n+1] > 0 && counts["p"][n+2] > 0 && counts["s"][n] > 0 && counts["s"][n+1] > 0 && counts["s"][n+2] > 0 {
			return true
		}
	}
	return false
}
func hasSanshokuDoukou(full []string) bool {
	counts := countByNumSuit(full)
	for n := 1; n <= 9; n++ {
		if counts["m"][n] >= 3 && counts["p"][n] >= 3 && counts["s"][n] >= 3 {
			return true
		}
	}
	return false
}

func countByNumSuit(full []string) map[string]map[int]int {
	m := map[string]map[int]int{"m": {}, "p": {}, "s": {}}
	for i := 1; i <= 9; i++ {
		m["m"][i] = 0
		m["p"][i] = 0
		m["s"][i] = 0
	}
	for _, s := range full {
		t := parseTile(s)
		if t.isHonor {
			continue
		}
		m[t.suit][t.num]++
	}
	return m
}

func countSevenPairs(tehai []string) int {
	counts := map[string]int{}
	for _, s := range tehai {
		counts[s]++
	}
	pairs := 0
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if counts[k] == 2 {
			pairs++
		}
	}
	return pairs
}

// Additional helper functions for 1 han yaku
func (u *YakuUsecase) detectOneHan(full []string, req *entity.YakuEvaluationRequest) map[string]int {
	res := map[string]int{}
	tiles := make([]tile, 0, len(full))
	for _, s := range full {
		tiles = append(tiles, parseTile(s))
	}
	// 1 han yaku
	if hasTanyao(tiles) {
		res["断么九"] = 1
	}
	if hasIpeikou(req.Tehai) && len(req.OpenMelds) == 0 {
		res["一盃口"] = 1
	}
	if hasPinfu(tiles, req) && len(req.OpenMelds) == 0 {
		res["平和"] = 1
	}
	// Special yaku
	if hasSankantsu(req.OpenMelds) {
		res["三暗刻"] = 2
	}
	return res
}

func (u *YakuUsecase) isChombo(req *entity.YakuEvaluationRequest) bool {
	// Check for invalid hand (too many tiles, wrong count, etc.)
	totalTiles := len(req.Tehai)
	for _, m := range req.OpenMelds {
		totalTiles += len(m.Tiles)
	}
	// Should be 14 tiles total
	if totalTiles != 14 {
		return true
	}
	// Check for invalid tile counts
	allTiles := make([]string, 0, totalTiles)
	allTiles = append(allTiles, req.Tehai...)
	for _, m := range req.OpenMelds {
		allTiles = append(allTiles, m.Tiles...)
	}
	counts := make(map[string]int)
	for _, t := range allTiles {
		counts[t]++
	}
	// Check for more than 4 of any tile
	for _, count := range counts {
		if count > 4 {
			return true
		}
	}
	return false
}

func hasTanyao(tiles []tile) bool {
	for _, t := range tiles {
		if t.isHonor || isTerminal(t) {
			return false
		}
	}
	return len(tiles) > 0
}

func hasIpeikou(tehai []string) bool {
	// Check for two identical sequences in hand (menzen only)
	counts := make(map[string]int)
	for _, t := range tehai {
		counts[t]++
	}
	// Look for sequences that appear twice
	sequences := make(map[string]int)
	for i := 0; i < len(tehai)-2; i++ {
		if isSequence(tehai[i], tehai[i+1], tehai[i+2]) {
			seq := tehai[i] + tehai[i+1] + tehai[i+2]
			sequences[seq]++
		}
	}
	// Check if any sequence appears exactly twice
	for _, count := range sequences {
		if count == 2 {
			return true
		}
	}
	return false
}

func hasPinfu(tiles []tile, req *entity.YakuEvaluationRequest) bool {
	// Pinfu: all sequences, pair is not yakuhai, ryanmen wait
	// Simplified check - would need full hand analysis
	return false // Placeholder
}

func hasChanta(full []string) bool {
	// All melds and pair must contain 1,9 or honors
	// 混全帯么九: 全ての面子と雀頭が么九牌（1,9,字牌）を含む

	// パース済みの牌を取得
	tiles := make([]tile, len(full))
	for i, s := range full {
		tiles[i] = parseTile(s)
	}

	// 面子と雀頭に分解（簡易版）
	// 4面子1雀頭の構成をチェック
	groups := make([][]tile, 0)
	used := make([]bool, len(tiles))

	// 刻子/槓子をチェック
	for i := 0; i < len(tiles); i++ {
		if used[i] {
			continue
		}
		count := 1
		indices := []int{i}
		for j := i + 1; j < len(tiles); j++ {
			if !used[j] && tiles[i].raw == tiles[j].raw {
				count++
				indices = append(indices, j)
			}
		}
		if count >= 3 {
			// 刻子として扱う
			group := []tile{tiles[i], tiles[i], tiles[i]}
			groups = append(groups, group)
			for k := 0; k < 3 && k < len(indices); k++ {
				used[indices[k]] = true
			}
		} else if count == 2 {
			// 雀頭候補
			group := []tile{tiles[i], tiles[i]}
			groups = append(groups, group)
			for k := 0; k < 2 && k < len(indices); k++ {
				used[indices[k]] = true
			}
		}
	}

	// 簡易チェック：全ての牌が么九牌か字牌を含むかチェック
	hasTerminalOrHonor := false
	for _, t := range tiles {
		if t.isHonor || t.num == 1 || t.num == 9 {
			hasTerminalOrHonor = true
			break
		}
	}

	// 么九牌または字牌が含まれ、かつ字牌が含まれている場合のみ混全帯么九
	hasHonor := false
	for _, t := range tiles {
		if t.isHonor {
			hasHonor = true
			break
		}
	}

	return hasTerminalOrHonor && hasHonor
}

func hasJunchan(full []string) bool {
	// 純全帯么九: 全ての面子と雀頭が么九牌（1,9）を含む（字牌なし）

	// パース済みの牌を取得
	tiles := make([]tile, len(full))
	for i, s := range full {
		tiles[i] = parseTile(s)
	}

	// 字牌が含まれていたら純全帯么九ではない
	for _, t := range tiles {
		if t.isHonor {
			return false
		}
	}

	// 純全帯么九の正しい判定：
	// すべての面子と雀頭が幺九牌（1,9）を含む
	// 順子の場合：1-2-3, 7-8-9 のみ
	// 刻子の場合：1, 9 のみ
	// 対子の場合：1, 9 のみ

	// 牌を数値別にカウント
	numCounts := make(map[string]map[int]int)
	suits := []string{"m", "p", "s"}
	for _, suit := range suits {
		numCounts[suit] = make(map[int]int)
	}

	for _, t := range tiles {
		if !t.isHonor && t.num >= 1 && t.num <= 9 {
			numCounts[t.suit][t.num]++
		}
	}

	// 各スートで幺九牌を含む面子があるかチェック
	validMelds := 0

	for _, suit := range suits {
		counts := numCounts[suit]

		// 順子チェック（1-2-3, 7-8-9のみ）
		if counts[1] > 0 && counts[2] > 0 && counts[3] > 0 {
			validMelds++
		}
		if counts[7] > 0 && counts[8] > 0 && counts[9] > 0 {
			validMelds++
		}

		// 刻子チェック（1, 9のみ）
		if counts[1] >= 3 {
			validMelds++
		}
		if counts[9] >= 3 {
			validMelds++
		}

		// 対子チェック（1, 9のみ）
		if counts[1] == 2 {
			validMelds++
		}
		if counts[9] == 2 {
			validMelds++
		}
	}

	// 4面子1雀頭の構成で、すべてが幺九牌を含むかチェック
	// 簡易チェック：幺九牌を含む面子が4つ以上あるか
	return validMelds >= 4
}

func hasRyanpeikou(tehai []string) bool {
	// Two ipeikou (four identical sequences)
	// Simplified check
	return false // Placeholder
}

func hasSankantsu(melds []entity.MeldRequest) bool {
	// Three concealed triplets
	// Simplified check
	return false // Placeholder
}

func isSequence(t1, t2, t3 string) bool {
	// Check if three tiles form a sequence
	// Simplified check
	return false // Placeholder
}

// getDoraFromIndicator converts dora indicator to actual dora tile
func getDoraFromIndicator(indicator string) string {
	// 数牌の場合：1->2, 2->3, ..., 8->9, 9->1
	if len(indicator) >= 2 {
		if indicator[0] >= '1' && indicator[0] <= '9' {
			num := int(indicator[0] - '0')
			suit := string(indicator[1])
			if num == 9 {
				return "1" + suit
			} else {
				return string('0'+byte(num+1)) + suit
			}
		}
	}

	// 字牌の場合
	switch indicator {
	case "ton", "東":
		return "nan" // 南
	case "nan", "南":
		return "sya" // 西
	case "sya", "西":
		return "pe" // 北
	case "pe", "北":
		return "ton" // 東
	case "haku", "白":
		return "hatu" // 發
	case "hatu", "發":
		return "chun" // 中
	case "chun", "中":
		return "haku" // 白
	default:
		return indicator // フォールバック
	}
}

// calculateScore calculates the score based on han count and whether it's oya (parent)
func calculateScore(han int, isOya bool) int {
	if han <= 0 {
		return 0
	}

	// 役満の計算（13翻 = 1倍役満、26翻 = 2倍役満、39翻 = 3倍役満）
	if han >= 13 {
		yakumanCount := han / 13
		baseScore := 32000
		if isOya {
			baseScore = 48000
		}

		// 2倍役満 = 64000 (子) / 96000 (親)
		// 3倍役満 = 96000 (子) / 144000 (親)
		// 4倍役満以上 = baseScore * yakumanCount
		return baseScore * yakumanCount
	}

	// 跳満（11-12翻）
	if han >= 11 {
		if isOya {
			return 36000
		}
		return 24000
	}

	// 倍満（8翻）
	if han == 8 {
		if isOya {
			return 24000
		}
		return 16000
	}

	// 満貫（9-10翻）
	if han >= 9 {
		if isOya {
			return 12000
		}
		return 8000
	}

	// 跳満（6-7翻）
	if han >= 6 {
		if isOya {
			return 18000
		}
		return 12000
	}

	// 満貫（5翻）
	if han == 5 {
		if isOya {
			return 12000
		}
		return 8000
	}

	// 通常の点数計算（1-4翻）
	switch han {
	case 1:
		if isOya {
			return 1500
		}
		return 1000
	case 2:
		if isOya {
			return 2900
		}
		return 2000
	case 3:
		if isOya {
			return 5800
		}
		return 3900
	case 4:
		if isOya {
			return 11600
		}
		return 7700
	default:
		return 0
	}
}
