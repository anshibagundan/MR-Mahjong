package entity

// 麻雀牌構造体
type Tile string

const (
	// 萬子（3人麻雀用）
	Tile1m Tile = "1m"
	Tile9m Tile = "9m"

	// 筒子
	Tile1p Tile = "1p"
	Tile2p Tile = "2p"
	Tile3p Tile = "3p"
	Tile4p Tile = "4p"
	Tile5p Tile = "5p"
	Tile6p Tile = "6p"
	Tile7p Tile = "7p"
	Tile8p Tile = "8p"
	Tile9p Tile = "9p"

	// 索子
	Tile1s Tile = "1s"
	Tile2s Tile = "2s"
	Tile3s Tile = "3s"
	Tile4s Tile = "4s"
	Tile5s Tile = "5s"
	Tile6s Tile = "6s"
	Tile7s Tile = "7s"
	Tile8s Tile = "8s"
	Tile9s Tile = "9s"

	// 字牌
	TileTon  Tile = "ton"  // 東
	TileNan  Tile = "nan"  // 南
	TileSya  Tile = "sya"  // 西
	TilePe   Tile = "pe"   // 北
	TileHaku Tile = "haku" // 白
	TileHatu Tile = "hatu" // 發
	TileChun Tile = "chun" // 中

	// 赤ドラ
	Tile5pr Tile = "5pr" // 筒子赤ドラ
	Tile5sr Tile = "5sr" // 索子赤ドラ
)

const (
	totalTilesThreePlayer = 108
)

// ゲーム用全牌を取得（3人麻雀、計108枚）
func GetAllTiles() []Tile {
	tiles := make([]Tile, 0, totalTilesThreePlayer)

	// 萬子（1mと9mのみ、各4枚）
	for _, tile := range []Tile{Tile1m, Tile9m} {
		for i := 0; i < 4; i++ {
			tiles = append(tiles, tile)
		}
	}

	// 筒子（通常5p×3 + 赤5pr×1、他4枚ずつ）
	for _, tile := range []Tile{Tile1p, Tile2p, Tile3p, Tile4p, Tile6p, Tile7p, Tile8p, Tile9p} {
		for i := 0; i < 4; i++ {
			tiles = append(tiles, tile)
		}
	}
	// 5p：通常3枚 + 赤1枚
	for i := 0; i < 3; i++ {
		tiles = append(tiles, Tile5p)
	}
	tiles = append(tiles, Tile5pr)

	// 索子（通常5s×3 + 赤5sr×1、他4枚ずつ）
	for _, tile := range []Tile{Tile1s, Tile2s, Tile3s, Tile4s, Tile6s, Tile7s, Tile8s, Tile9s} {
		for i := 0; i < 4; i++ {
			tiles = append(tiles, tile)
		}
	}
	// 5s：通常3枚 + 赤1枚
	for i := 0; i < 3; i++ {
		tiles = append(tiles, Tile5s)
	}
	tiles = append(tiles, Tile5sr)

	// 字牌（各4枚）
	for _, tile := range []Tile{TileTon, TileNan, TileSya, TilePe, TileHaku, TileHatu, TileChun} {
		for i := 0; i < 4; i++ {
			tiles = append(tiles, tile)
		}
	}

	return tiles
}
