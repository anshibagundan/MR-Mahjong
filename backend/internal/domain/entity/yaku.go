package entity

// MeldRequest REST request meld representation
type MeldRequest struct {
	Type  string   `json:"type"` // pon, chi, kan
	Tiles []string `json:"tiles"`
}

// YakuEvaluationRequest REST request body for yaku evaluation
type YakuEvaluationRequest struct {
	Tehai             []string      `json:"tehai"`
	OpenMelds         []MeldRequest `json:"openMelds"`
	WinTile           string        `json:"winTile"`
	IsTsumo           bool          `json:"isTsumo"`
	Riichi            bool          `json:"riichi"`
	Ippatsu           bool          `json:"ippatsu"`
	DoraIndicators    []string      `json:"doraIndicators"`
	UraDoraIndicators []string      `json:"uraDoraIndicators"`
	RoundWind         string        `json:"roundWind"` // east, south, west, north
	SeatWind          string        `json:"seatWind"`  // east, south, west, north
	Tenhou            bool          `json:"tenhou"`    // 天和
	Chiihou           bool          `json:"chiihou"`   // 地和
	Renhou            bool          `json:"renhou"`    // 人和
}

// YakuItem represents a identified yaku
type YakuItem struct {
	Name string `json:"name"`
	Han  int    `json:"han"`
}

// YakuEvaluationResponse REST response body for yaku evaluation
type YakuEvaluationResponse struct {
	Yaku         []YakuItem `json:"yaku"`
	Fu           int        `json:"fu"`
	Han          int        `json:"han"`
	DoraCount    int        `json:"doraCount"`
	UraDoraCount int        `json:"uraDoraCount"`
	TotalHan     int        `json:"totalHan"`
	Yakuman      []string   `json:"yakuman"`
	Score        int        `json:"score"`
	IsChombo     bool       `json:"isChombo"`
}
