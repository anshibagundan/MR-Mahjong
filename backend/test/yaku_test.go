package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mahjong-backend/internal/domain/entity"
	"mahjong-backend/internal/interface/handler"
	"mahjong-backend/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestYakuEvaluation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	yakuUC := usecase.NewYakuUsecase()
	yakuHandler := handler.NewYakuHandler(yakuUC)

	router := gin.New()
	router.POST("/api/v1/yaku", yakuHandler.HandleEvaluateYaku)

	testCases := []struct {
		name              string
		request           entity.YakuEvaluationRequest
		expectedScore     int
		expectedTotalHan  int
		expectedIsChombo  bool
		expectedDoraCount int
		containsYaku      []string
	}{
		{
			name: "1翻 - 立直_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "2m", "3m", "4m", "5m", "6m", "1s", "2s", "3s", "2p", "3p", "4p", "5p", "6p"},
				WinTile:   "6p",
				Riichi:    true,
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    1000,
			expectedTotalHan: 1,
			containsYaku:     []string{"立直"},
		},
		{
			name: "2翻 - 立直_門前清自摸和_ツモ",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "2m", "3m", "4m", "5m", "6m", "1s", "2s", "3s", "2p", "3p", "4p", "5p", "6p"},
				WinTile:   "6p",
				Riichi:    true,
				IsTsumo:   true,
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    2000,
			expectedTotalHan: 2, // 立直(1翻) + 門前清自摸和(1翻) = 2翻
			containsYaku:     []string{"立直", "門前清自摸和"},
		},
		{
			name: "2翻 - 立直_断么九_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"2m", "3m", "4m", "5m", "6m", "7m", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "8p"},
				WinTile:   "2p",
				Riichi:    true,
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    2000,
			expectedTotalHan: 2,
			containsYaku:     []string{"立直", "断么九"},
		},
		{
			name: "2翻 - 門前清自摸和_断么九_門前_ツモ",
			request: entity.YakuEvaluationRequest{
				Tehai:             []string{"2m", "3m", "4m", "5m", "6m", "7m", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "8p"},
				OpenMelds:         []entity.MeldRequest{},
				WinTile:           "8p",
				IsTsumo:           true,
				Riichi:            false,
				Ippatsu:           false,
				DoraIndicators:    []string{},
				UraDoraIndicators: []string{},
				RoundWind:         "east",
				SeatWind:          "south",
			},
			expectedScore:    2000,
			expectedTotalHan: 2,
			containsYaku:     []string{"門前清自摸和", "断么九"},
		},
		{
			name: "3翻 - 立直_一気通貫_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "2m", "3m", "4m", "5m", "6m", "7m", "8m", "9m", "2p", "3p", "4p", "5p", "6p"},
				WinTile:   "6p",
				Riichi:    true,
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    3900,
			expectedTotalHan: 3,
			containsYaku:     []string{"立直", "一気通貫"},
		},
		{
			name: "4翻 - 対々和_役牌(白)_混全帯么九_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"2m", "2m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"haku", "haku", "haku"}},
					{Type: "pon", Tiles: []string{"3p", "3p", "3p"}},
					{Type: "pon", Tiles: []string{"4s", "4s", "4s"}},
					{Type: "pon", Tiles: []string{"5s", "5s", "5s"}},
				},
				WinTile:   "2m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    7700,
			expectedTotalHan: 4,
			containsYaku:     []string{"対々和", "混全帯么九", "役牌(白)"},
		},
		{
			name: "2翻 - 立直_断么九_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"2m", "3m", "4m", "5m", "6m", "7m", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "8p"},
				WinTile:   "8p",
				Riichi:    true,
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    2000,
			expectedTotalHan: 2,
			containsYaku:     []string{"立直", "断么九"},
		},
		{
			name: "1翻 - 断么九_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"2m", "3m", "4m", "5m", "6m", "7m", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "8p"},
				WinTile:   "8p",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    1000,
			expectedTotalHan: 1,
			containsYaku:     []string{"断么九"},
		},
		{
			name: "3翻 - 七対子_断么九_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"2m", "2m", "3m", "3m", "4m", "4m", "5m", "5m", "6m", "6m", "7m", "7m", "8p", "8p"},
				WinTile:   "8p",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    3900,
			expectedTotalHan: 3,
			containsYaku:     []string{"七対子", "断么九"},
		},
		{
			name: "8翻 - 対々和_清一色_断么九_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"2m", "2m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"3m", "3m", "3m"}},
					{Type: "pon", Tiles: []string{"4m", "4m", "4m"}},
					{Type: "pon", Tiles: []string{"5m", "5m", "5m"}},
					{Type: "pon", Tiles: []string{"6m", "6m", "6m"}},
				},
				WinTile:   "2m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    16000,
			expectedTotalHan: 8,
			containsYaku:     []string{"対々和", "清一色", "断么九"},
		},
		{
			name: "6翻 - 対々和_混一色_役牌(白)_混全帯么九_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"1m", "1m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"haku", "haku", "haku"}},
					{Type: "pon", Tiles: []string{"2m", "2m", "2m"}},
					{Type: "pon", Tiles: []string{"3m", "3m", "3m"}},
					{Type: "pon", Tiles: []string{"4m", "4m", "4m"}},
				},
				WinTile:   "1m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    12000,
			expectedTotalHan: 6,
			containsYaku:     []string{"対々和", "混一色", "混全帯么九", "役牌(白)"},
		},
		{
			name: "7翻 - 対々和_清一色_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"5m", "5m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"1m", "1m", "1m"}},
					{Type: "pon", Tiles: []string{"2m", "2m", "2m"}},
					{Type: "pon", Tiles: []string{"3m", "3m", "3m"}},
					{Type: "pon", Tiles: []string{"4m", "4m", "4m"}},
				},
				WinTile:   "5m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    12000,
			expectedTotalHan: 7, // 清一色(5翻) + 対々和(2翻) = 7翻
			containsYaku:     []string{"清一色", "対々和"},
		},
		{
			name: "6翻 - 対々和_役牌(白)_混全帯么九_ドラ2_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"2m", "2m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"haku", "haku", "haku"}},
					{Type: "pon", Tiles: []string{"3p", "3p", "3p"}},
					{Type: "pon", Tiles: []string{"4s", "4s", "4s"}},
					{Type: "pon", Tiles: []string{"5s", "5s", "5s"}},
				},
				WinTile:        "2m",
				DoraIndicators: []string{"1m"},
				RoundWind:      "east",
				SeatWind:       "south",
			},
			expectedScore:     12000,
			expectedTotalHan:  6,
			expectedDoraCount: 2,
			containsYaku:      []string{"対々和", "混全帯么九", "役牌(白)"},
		},
		{
			name: "役満 - 国士無双十三面待ち_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "9m", "1p", "9p", "1s", "9s", "ton", "nan", "sya", "pe", "haku", "hatu", "chun", "chun"},
				WinTile:   "chun",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    32000,
			expectedTotalHan: 13,
			containsYaku:     []string{},
		},
		{
			name: "役満 - 大四喜_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"ton", "ton", "ton", "nan", "nan", "nan", "sya", "sya", "sya", "pe", "pe", "pe", "1m", "1m"},
				WinTile:   "1m",
				IsTsumo:   false,
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    32000,
			expectedTotalHan: 13,
			containsYaku:     []string{},
		},

		{
			name: "6翻 - 対々和_役牌(場風)_混全帯么九_門清_ツモ",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "1m", "1m", "2p", "2p", "2p", "3s", "3s", "3s", "ton", "ton", "ton", "haku", "haku"},
				WinTile:   "haku",
				IsTsumo:   true,
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    12000,
			expectedTotalHan: 6, // 門清(1翻) + 対々和(2翻) + 混全帯么九(2翻) + 役牌場風(1翻) = 6翻
			containsYaku:     []string{"門前清自摸和", "対々和", "混全帯么九", "役牌(場風)"},
		},
		{
			name: "役満 - 大三元_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"1m", "1m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"haku", "haku", "haku"}},
					{Type: "pon", Tiles: []string{"hatu", "hatu", "hatu"}},
					{Type: "pon", Tiles: []string{"chun", "chun", "chun"}},
					{Type: "pon", Tiles: []string{"2m", "2m", "2m"}},
				},
				WinTile:   "1m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    32000,
			expectedTotalHan: 13,
			containsYaku:     []string{},
		},
		{
			name: "役満 - 緑一色_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"2s", "2s", "2s", "3s", "3s", "3s", "4s", "4s", "4s", "6s", "6s", "6s", "hatu", "hatu"},
				WinTile:   "hatu",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    32000,
			expectedTotalHan: 13, // 緑一色(13翻)
			containsYaku:     []string{},
		},
		{
			name: "2倍役満 - 大四喜_字一色_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"ton", "ton", "ton", "nan", "nan", "nan", "sya", "sya", "sya", "pe", "pe", "pe", "haku", "haku"},
				WinTile:   "haku",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    64000,
			expectedTotalHan: 26,
			containsYaku:     []string{},
		},
		{
			name: "役満 - 清老頭_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "1m", "1m", "9m", "9m", "9m", "1p", "1p", "1p", "9p", "9p", "9p", "1s", "1s"},
				WinTile:   "1s",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    32000,
			expectedTotalHan: 13,
			containsYaku:     []string{},
		},
		{
			name: "4翻 - 役牌(東)_場風_対々和_混全帯么九_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"2m", "2m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"ton", "ton", "ton"}},
					{Type: "pon", Tiles: []string{"3p", "3p", "3p"}},
					{Type: "pon", Tiles: []string{"4s", "4s", "4s"}},
					{Type: "pon", Tiles: []string{"5s", "5s", "5s"}},
				},
				WinTile:   "2m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    7700,
			expectedTotalHan: 4,
			containsYaku:     []string{"混全帯么九", "対々和", "役牌(場風)"},
		},
		{
			name: "4翻 - 役牌(南)_自風_対々和_混全帯么九_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"3m", "3m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"nan", "nan", "nan"}},
					{Type: "pon", Tiles: []string{"4p", "4p", "4p"}},
					{Type: "pon", Tiles: []string{"5s", "5s", "5s"}},
					{Type: "pon", Tiles: []string{"6s", "6s", "6s"}},
				},
				WinTile:   "3m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    7700,
			expectedTotalHan: 4,
			containsYaku:     []string{"対々和", "混全帯么九", "役牌(自風)"},
		},
		{
			name: "9翻 - 清一色_七対子_断么九_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"2m", "2m", "3m", "3m", "4m", "4m", "5m", "5m", "6m", "6m", "7m", "7m", "8m", "8m"},
				WinTile:   "8m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    8000, // 子の9翻 = 満貫8000点
			expectedTotalHan: 9,    // 清一色(6翻) + 七対子(2翻) + 断么九(1翻)
			containsYaku:     []string{"清一色", "七対子", "断么九"},
		},
		{
			name: "9翻 - 清一色_七対子_断么九_門前_ロン_親",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"2m", "2m", "3m", "3m", "4m", "4m", "5m", "5m", "6m", "6m", "7m", "7m", "8m", "8m"},
				WinTile:   "8m",
				RoundWind: "east",
				SeatWind:  "east", // 親（場風と自風が同じ）
			},
			expectedScore:    12000, // 親の9翻 = 満貫12000点
			expectedTotalHan: 9,     // 清一色(6翻) + 七対子(2翻) + 断么九(1翻)
			containsYaku:     []string{"清一色", "七対子", "断么九"},
		},
		{
			name: "8翻 - 清一色_対々和_断么九_鳴きあり_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"5m", "5m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"2m", "2m", "2m"}},
					{Type: "pon", Tiles: []string{"3m", "3m", "3m"}},
					{Type: "pon", Tiles: []string{"4m", "4m", "4m"}},
					{Type: "pon", Tiles: []string{"6m", "6m", "6m"}},
				},
				WinTile:   "5m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    16000,
			expectedTotalHan: 8, // 清一色(5翻) + 対々和(2翻) + 断么九(1翻)
			containsYaku:     []string{"清一色", "対々和", "断么九"},
		},
		{
			name: "7翻 - 混一色_七対子_混全帯么九_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"2m", "2m", "3m", "3m", "4m", "4m", "5m", "5m", "6m", "6m", "ton", "ton", "nan", "nan"},
				WinTile:   "nan",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    12000,
			expectedTotalHan: 7, // 混一色(3翻) + 七対子(2翻) + 混全帯么九(2翻)
			containsYaku:     []string{"混一色", "七対子", "混全帯么九"},
		},
		{
			name: "6翻 - 混一色_対々和_混全帯么九_役牌鳴きあり_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"ton", "ton"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"2m", "2m", "2m"}},
					{Type: "pon", Tiles: []string{"3m", "3m", "3m"}},
					{Type: "pon", Tiles: []string{"4m", "4m", "4m"}},
					{Type: "pon", Tiles: []string{"nan", "nan", "nan"}},
				},
				WinTile:   "ton",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    12000,
			expectedTotalHan: 6, // 混一色(2翻) + 対々和(2翻) + 混全帯么九(1翻) + 役牌自風(1翻)
			containsYaku:     []string{"混一色", "対々和", "混全帯么九", "役牌(自風)"},
		},
		{
			name: "3翻 - 純全帯么九_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "2m", "3m", "7m", "8m", "9m", "1p", "2p", "3p", "7p", "8p", "9p", "1s", "1s"},
				WinTile:   "1s",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    3900,
			expectedTotalHan: 3,
			containsYaku:     []string{"純全帯么九"},
		},
		{
			name: "2翻 - 純全帯么九_鳴きあり_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"1s", "1s"},
				OpenMelds: []entity.MeldRequest{
					{Type: "chi", Tiles: []string{"1m", "2m", "3m"}},
					{Type: "chi", Tiles: []string{"7m", "8m", "9m"}},
					{Type: "chi", Tiles: []string{"1p", "2p", "3p"}},
					{Type: "chi", Tiles: []string{"7p", "8p", "9p"}},
				},
				WinTile:   "1s",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    2000,
			expectedTotalHan: 2,
			containsYaku:     []string{"純全帯么九"},
		},
		{
			name: "2翻 - 混全帯么九_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "2m", "3m", "7m", "8m", "9m", "1p", "2p", "3p", "7p", "8p", "9p", "ton", "ton"},
				WinTile:   "ton",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    2000,
			expectedTotalHan: 2,
			containsYaku:     []string{"混全帯么九"},
		},
		{
			name: "1翻 - 混全帯么九_鳴きあり_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"ton", "ton"},
				OpenMelds: []entity.MeldRequest{
					{Type: "chi", Tiles: []string{"1m", "2m", "3m"}},
					{Type: "chi", Tiles: []string{"7m", "8m", "9m"}},
					{Type: "chi", Tiles: []string{"1p", "2p", "3p"}},
					{Type: "chi", Tiles: []string{"7p", "8p", "9p"}},
				},
				WinTile:   "ton",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    1000,
			expectedTotalHan: 1,
			containsYaku:     []string{"混全帯么九"},
		},
		{
			name: "3翻 - 純全帯么九_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "2m", "3m", "7m", "8m", "9m", "1p", "2p", "3p", "7p", "8p", "9p", "1s", "1s"},
				WinTile:   "1s",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    3900,
			expectedTotalHan: 3,
			containsYaku:     []string{"純全帯么九"},
		},
		{
			name: "2翻 - 純全帯么九_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"1s", "1s"},
				OpenMelds: []entity.MeldRequest{
					{Type: "chi", Tiles: []string{"1m", "2m", "3m"}},
					{Type: "chi", Tiles: []string{"7m", "8m", "9m"}},
					{Type: "chi", Tiles: []string{"1p", "2p", "3p"}},
					{Type: "chi", Tiles: []string{"7p", "8p", "9p"}},
				},
				WinTile:   "1s",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    2000,
			expectedTotalHan: 2,
			containsYaku:     []string{"純全帯么九"},
		},
		{
			name: "5翻 - 三色同順_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "2m", "3m", "1p", "2p", "3p", "1s", "2s", "3s", "4s", "4s", "5s", "5s", "5s"},
				WinTile:   "5s",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    8000,
			expectedTotalHan: 5,
			containsYaku:     []string{"三色同順"},
		},
		{
			name: "3翻 - 三色同順_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"1s", "2s", "3s", "5s", "5s"},
				OpenMelds: []entity.MeldRequest{
					{Type: "chi", Tiles: []string{"1m", "2m", "3m"}},
					{Type: "chi", Tiles: []string{"1p", "2p", "3p"}},
					{Type: "pon", Tiles: []string{"4s", "4s", "4s"}},
				},
				WinTile:   "5s",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    3900,
			expectedTotalHan: 3,
			containsYaku:     []string{"三色同順"},
		},
		{
			name: "5翻 - 三色同刻_対々和_門前_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"2m", "2m", "2m", "2p", "2p", "2p", "2s", "2s", "2s", "3m", "3m", "4m", "4m", "4m"},
				WinTile:   "4m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    8000,
			expectedTotalHan: 5,
			containsYaku:     []string{"三色同刻", "対々和", "断么九"},
		},
		{
			name: "5翻 - 三色同刻_対々和_鳴きあり_ロン",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"2s", "2s", "2s", "4m", "4m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"2m", "2m", "2m"}},
					{Type: "pon", Tiles: []string{"2p", "2p", "2p"}},
					{Type: "pon", Tiles: []string{"3m", "3m", "3m"}},
				},
				WinTile:   "4m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    8000,
			expectedTotalHan: 5,
			containsYaku:     []string{"三色同刻", "対々和", "断么九"},
		},
		{
			name: "チョンボ - 牌数不正（手牌不足）",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "2m", "3m"},
				WinTile:   "4m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    0,
			expectedTotalHan: 0,
			expectedIsChombo: true,
		},
		{
			name: "チョンボ - 牌数不正（手牌過多）",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "2m", "3m", "4m", "5m", "6m", "7m", "8m", "9m", "1p", "2p", "3p", "4p", "5p", "6p"},
				WinTile:   "7p",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    0,
			expectedTotalHan: 0,
			expectedIsChombo: true,
		},
		{
			name: "チョンボ - 同一牌5枚以上",
			request: entity.YakuEvaluationRequest{
				Tehai:     []string{"1m", "1m", "1m", "1m", "1m", "2m", "3m", "4m", "5m", "6m", "7m", "8m", "9m", "9m"},
				WinTile:   "9m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    0,
			expectedTotalHan: 0,
			expectedIsChombo: true,
		},
		{
			name: "チョンボ - 副露で同一牌5枚以上",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"1m", "1m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"2m", "2m", "2m"}},
					{Type: "pon", Tiles: []string{"2m", "2m", "2m"}}, // 同一牌6枚
					{Type: "pon", Tiles: []string{"3m", "3m", "3m"}},
					{Type: "pon", Tiles: []string{"4m", "4m", "4m"}},
				},
				WinTile:   "1m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    0,
			expectedTotalHan: 0,
			expectedIsChombo: true,
		},
		{
			name: "チョンボ - 手牌と副露で同一牌5枚以上",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"1m", "1m", "1m", "1m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"2m", "2m", "2m"}},
					{Type: "pon", Tiles: []string{"3m", "3m", "3m"}},
					{Type: "pon", Tiles: []string{"4m", "4m", "4m"}},
					{Type: "pon", Tiles: []string{"5m", "5m", "5m"}},
				},
				WinTile:   "1m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    0,
			expectedTotalHan: 0,
			expectedIsChombo: true,
		},
		{
			name: "チョンボ - 副露面子数過多",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"1m", "1m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"2m", "2m", "2m"}},
					{Type: "pon", Tiles: []string{"3m", "3m", "3m"}},
					{Type: "pon", Tiles: []string{"4m", "4m", "4m"}},
					{Type: "pon", Tiles: []string{"5m", "5m", "5m"}},
					{Type: "pon", Tiles: []string{"6m", "6m", "6m"}},
				},
				WinTile:   "1m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    0,
			expectedTotalHan: 0,
			expectedIsChombo: true,
		},
		{
			name: "チョンボ - 副露面子の牌数不正（刻子が2枚）",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "pon", Tiles: []string{"2m", "2m"}}, // 2枚しかない
				},
				WinTile:   "1m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    0,
			expectedTotalHan: 0,
			expectedIsChombo: true,
		},
		{
			name: "チョンボ - 副露面子の牌数不正（順子が2枚）",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "chi", Tiles: []string{"2m", "3m"}}, // 2枚しかない
				},
				WinTile:   "1m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    0,
			expectedTotalHan: 0,
			expectedIsChombo: true,
		},
		{
			name: "チョンボ - 副露面子の牌数不正（槓子が3枚）",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "kan", Tiles: []string{"2m", "2m", "2m"}}, // 3枚しかない
				},
				WinTile:   "1m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    0,
			expectedTotalHan: 0,
			expectedIsChombo: true,
		},
		{
			name: "チョンボ - 副露面子の牌数不正（槓子が5枚）",
			request: entity.YakuEvaluationRequest{
				Tehai: []string{"1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m", "1m"},
				OpenMelds: []entity.MeldRequest{
					{Type: "kan", Tiles: []string{"2m", "2m", "2m", "2m", "2m"}}, // 5枚
				},
				WinTile:   "1m",
				RoundWind: "east",
				SeatWind:  "south",
			},
			expectedScore:    0,
			expectedTotalHan: 0,
			expectedIsChombo: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/api/v1/yaku", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response entity.YakuEvaluationResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedIsChombo, response.IsChombo)
			assert.Equal(t, tc.expectedScore, response.Score)
			assert.Equal(t, tc.expectedTotalHan, response.TotalHan)

			if tc.expectedDoraCount > 0 {
				assert.Equal(t, tc.expectedDoraCount, response.DoraCount)
			}

			if !tc.expectedIsChombo && len(tc.containsYaku) > 0 {
				actualYakuNames := make([]string, len(response.Yaku))
				for i, yaku := range response.Yaku {
					actualYakuNames[i] = yaku.Name
				}

				for _, expectedName := range tc.containsYaku {
					found := false
					for _, actualName := range actualYakuNames {
						if actualName == expectedName {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected yaku '%s' not found in %v", expectedName, actualYakuNames)
				}
			}

			t.Logf("Response: %+v", response)
		})
	}
}
