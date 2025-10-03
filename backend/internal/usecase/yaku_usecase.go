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
	for _, y := range u.detectYakuman(fullTiles, req) {
		result.Yakuman = append(result.Yakuman, y)
		result.TotalHan += 13
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

	// Yakuhai example: 白/發/中 ポン面子 or 雀頭+面子
	// If openMelds contains pon with honor tiles, add 1 han for each type present
	honorMap := map[string]string{"白": "役牌(白)", "發": "役牌(發)", "中": "役牌(中)", "haku": "役牌(白)", "hatu": "役牌(發)", "chun": "役牌(中)"}
	// Winds (only if matches round or seat wind)
	windNames := map[string]string{"ton": "東", "nan": "南", "sya": "西", "pe": "北", "東": "東", "南": "南", "西": "西", "北": "北"}
	for _, m := range req.OpenMelds {
		if strings.ToLower(m.Type) == "pon" && len(m.Tiles) == 3 {
			tile0 := m.Tiles[0]
			// 三元牌
			if name, ok := honorMap[tile0]; ok {
				result.Yaku = append(result.Yaku, entity.YakuItem{Name: name, Han: 1})
				result.Han += 1
				continue
			}
			// 風牌（場風/自風）
			if w, ok := windNames[tile0]; ok {
				if matchesWind(w, req.RoundWind) {
					result.Yaku = append(result.Yaku, entity.YakuItem{Name: "役牌(場風)", Han: 1})
					result.Han += 1
				}
				if matchesWind(w, req.SeatWind) {
					result.Yaku = append(result.Yaku, entity.YakuItem{Name: "役牌(自風)", Han: 1})
					result.Han += 1
				}
			}
		}
	}

	// Dora counting (very simplified): count exact matches of indicators in tehai
	indicatorSet := make(map[string]int)
	for _, d := range req.DoraIndicators {
		indicatorSet[d]++
	}
	for _, t := range req.Tehai {
		if indicatorSet[t] > 0 {
			result.DoraCount++
		}
	}
	uraSet := make(map[string]int)
	for _, d := range req.UraDoraIndicators {
		uraSet[d]++
	}
	for _, t := range req.Tehai {
		if uraSet[t] > 0 {
			result.UraDoraCount++
		}
	}

	result.TotalHan = result.Han + result.DoraCount + result.UraDoraCount
	// Fu計算は不要なので、総翻のみで点数算出（暫定: 1翻=400点）
	result.Score = result.TotalHan * 400
	return result
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
	// Kokushi
	if isKokushi(full) {
		return []string{"国士無双"}
	}
	// Ryuuiisou
	if isRyuuiisou(tiles) {
		return []string{"緑一色"}
	}
	// Tsuuiisou
	if allHonors(tiles) {
		return []string{"字一色"}
	}
	// Chinroutou
	if allTerminals(tiles) {
		return []string{"清老頭"}
	}
	// Daisangen / Shousuushi / Daisuushi
	dragonTriplets := countTripletKinds(full, []string{"白", "haku", "發", "hatu", "中", "chun"})
	if dragonTriplets >= 3 {
		return []string{"大三元"}
	}
	windTriplets := countTripletKinds(full, []string{"ton", "東", "nan", "南", "sya", "西", "pe", "北"})
	if windTriplets == 4 {
		return []string{"大四喜"}
	}
	if windTriplets == 3 && hasPair(full, []string{"ton", "東", "nan", "南", "sya", "西", "pe", "北"}) {
		return []string{"小四喜"}
	}
	return nil
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
	// Simplified check
	return false // Placeholder
}

func hasJunchan(full []string) bool {
	// All melds and pair must contain 1,9 (no honors)
	// Simplified check
	return false // Placeholder
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
