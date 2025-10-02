package usecase

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"mahjong-backend/internal/domain/entity"
)

// ゲーム業務ロジック（メモリ保存）
type GameUsecase struct {
	game    *entity.Game
	players map[string]bool // プレイヤーID -> 参加中フラグ
	mutex   sync.RWMutex
}

// GameUsecaseのインスタンス
func NewGameUsecase() *GameUsecase {
	return &GameUsecase{
		game:    entity.NewGame(),
		players: make(map[string]bool),
	}
}

// ゲームを取得
func (u *GameUsecase) GetGame() *entity.Game {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.game
}

// プレイヤーをゲームに追加
func (u *GameUsecase) AddPlayerToGame(playerID string) (*entity.Game, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// プレイヤーが既にゲームに参加済みかチェック
	if u.players[playerID] {
		// プレイヤーは既にこのゲームに参加済み
		return u.game, nil
	}

	if !u.game.AddPlayer(playerID) {
		return nil, errors.New("game is full")
	}

	u.players[playerID] = true
	return u.game, nil
}

// プレイヤーをゲームから削除
func (u *GameUsecase) RemovePlayerFromGame(playerID string) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	// ゲームからプレイヤーを削除
	u.game.RemovePlayer(playerID)

	// マッピングからプレイヤーを削除
	delete(u.players, playerID)

	// ゲームが空ならリセット
	if len(u.game.Players) == 0 {
		u.game = entity.NewGame()
	}

	return nil
}

// プレイヤーが参加中かチェック
func (u *GameUsecase) IsPlayerInGame(playerID string) bool {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.players[playerID]
}

// ゲーム初期化・開始（牌配布）
func (u *GameUsecase) StartGame(wsUsecase *WebSocketUsecase) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if !u.game.CanStart() {
		return nil // ゲームはまだ開始できない
	}

	// 牌をシャッフル
	tiles := entity.GetAllTiles()
	u.shuffleTiles(tiles)

	// 牌を配布
	u.distributeTiles(u.game, tiles)

	// ゲーム状態を更新
	u.game.Status = entity.GameStatusPlaying

	// 各プレイヤーに手牌を含むゲーム開始メッセージを送信
	for _, player := range u.game.Players {
		gameStartMessage := &entity.GameStartMessage{
			PlayerID: player.ID,
			Tehai:    player.Tehai,
			Wanpai:   u.game.Wanpai,
			Yama:     u.game.Yama,
			Players:  u.createPlayersViewForPlayer(u.game.Players),
		}

		message := &entity.WebSocketMessage{
			Type: entity.MessageTypeGameStart,
			Data: gameStartMessage,
		}

		if err := wsUsecase.SendToPlayer(player.ID, message); err != nil {
			return err
		}
	}

	return nil
}

// 牌配列をシャッフル
func (u *GameUsecase) shuffleTiles(tiles []entity.Tile) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(tiles), func(i, j int) {
		tiles[i], tiles[j] = tiles[j], tiles[i]
	})
}

// プレイヤー、王牌、山牌に牌を配布
func (u *GameUsecase) distributeTiles(game *entity.Game, tiles []entity.Tile) {
	index := 0

	// 親は14枚、子は13枚
	for _, player := range game.Players {
		tileCount := 13
		if player.IsHost {
			tileCount = 14 // 親は14枚
		}
		for i := 0; i < tileCount; i++ {
			player.Tehai = append(player.Tehai, tiles[index])
			index++
		}
	}

	// 王牌設定（計13枚）
	// 表ドラ1枚 + 裏ドラ4枚 + 嶺上牌8枚 = 13枚
	wanpaiTiles := tiles[index : index+13]
	index += 13

	game.Wanpai = &entity.Wanpai{
		RevealedDora:    []entity.Tile{wanpaiTiles[0]}, // 表ドラ1枚
		UnrevealedDoras: wanpaiTiles[1:5],              // 裏ドラ4枚
		Rinsyan:         wanpaiTiles[5:13],             // 嶺上牌8枚
	}

	// 残り牌は山牌へ（108 - (14 + 13 + 13) - 13 = 55枚）
	game.Yama = tiles[index:]
}

// 特定プレイヤー用のプレイヤー一覧を作成
// 透明性のため全プレイヤーの手牌を表示
func (u *GameUsecase) createPlayersViewForPlayer(allPlayers []*entity.Player) []*entity.Player {
	playersView := make([]*entity.Player, len(allPlayers))

	for i, player := range allPlayers {
		// 透明性のため全プレイヤーの手牌を表示
		playersView[i] = &entity.Player{
			ID:     player.ID,
			Tehai:  player.Tehai,
			IsHost: player.IsHost,
		}
	}

	return playersView
}
