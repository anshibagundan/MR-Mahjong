package entity

// ゲーム状態定数
const (
	GameStatusWaiting  = "waiting"
	GameStatusPlaying  = "playing"
	GameStatusFinished = "finished"
)

// 3人麻雀 固定値
const (
	ThreePlayerMahjongMaxPlayers      = 3
	InitialTileCapacityPerPlayerChild = 13
)

// プレイヤー構造体
type Player struct {
	ID     string `json:"id"`
	Tehai  []Tile `json:"tehai"`  // 手牌
	IsHost bool   `json:"isHost"` // 親かどうか
}

// 王牌構造体
type Wanpai struct {
	RevealedDora    []Tile `json:"revealedDora"`    // 表ドラ
	KanDoras        []Tile `json:"kanDoras"`        // カンドラ（カン成立時に順次公開）
	UnrevealedDoras []Tile `json:"unrevealedDoras"` // 裏ドラ
	Rinsyan         []Tile `json:"rinsyan"`         // 嶺上牌
}

// ゲーム構造体
type Game struct {
	Players    []*Player `json:"players"`
	Wanpai     *Wanpai   `json:"wanpai"`
	Yama       []Tile    `json:"yama"`       // 山牌
	Status     string    `json:"status"`     // ゲーム状態
	MaxPlayers int       `json:"maxPlayers"` // 最大プレイヤー数（3人麻雀）
}

// 新しいゲームを作成
func NewGame() *Game {
	return &Game{
		Players:    make([]*Player, 0, ThreePlayerMahjongMaxPlayers),
		Status:     GameStatusWaiting,
		MaxPlayers: ThreePlayerMahjongMaxPlayers,
	}
}

// プレイヤーをゲームに追加
func (g *Game) AddPlayer(playerID string) bool {
	if len(g.Players) >= g.MaxPlayers {
		return false
	}

	isHost := len(g.Players) == 0 // 最初のプレイヤーが親
	player := &Player{
		ID:     playerID,
		Tehai:  make([]Tile, 0, InitialTileCapacityPerPlayerChild),
		IsHost: isHost,
	}

	g.Players = append(g.Players, player)
	return true
}

// プレイヤーをゲームから削除
func (g *Game) RemovePlayer(playerID string) bool {
	for i, player := range g.Players {
		if player.ID == playerID {
			// プレイヤーをスライスから削除
			g.Players = append(g.Players[:i], g.Players[i+1:]...)

			// 削除されたプレイヤーが親の場合、残りの最初のプレイヤーを親にする
			if player.IsHost && len(g.Players) > 0 {
				g.Players[0].IsHost = true
			}

			return true
		}
	}
	return false
}

// ゲーム開始可能かどうか
func (g *Game) CanStart() bool {
	return len(g.Players) == g.MaxPlayers && g.Status == GameStatusWaiting
}

// IDでプレイヤーを取得
func (g *Game) GetPlayerByID(playerID string) *Player {
	for _, player := range g.Players {
		if player.ID == playerID {
			return player
		}
	}
	return nil
}
