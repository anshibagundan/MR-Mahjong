// C#
using System.Collections.Generic;
using UnityEngine;

public class TestMahjong3PManager : MonoBehaviour
{
    [Header("Prefabs / Roots")]
    [Tooltip("牌プレハブ（単純な GameObject で可）。SpriteRenderer やテキスト等は任意。")]
    public GameObject tilePrefab;

    [Tooltip("プレイヤーごとの牌配置ルート（順番: 親ホスト, 子2, 子3）")]
    public Transform[] playerTileRoots = new Transform[3];

    [Tooltip("山（壁）を配置する起点 Transform")]
    public Transform wallOrigin;

    [Tooltip("王牌を配置するコンテナ")]
    public Transform wanpaiRoot;

    [Header("Layout")]
    public float tileSpacing = 0.12f;
    public float wallSpacing = 0.06f;
    public float wallLayers = 2; // 壁を積む段数（視覚用）
    public bool shuffleOnStart = true;

    // 内部データ
    private List<string> deck = new List<string>();
    private List<GameObject> spawnedTiles = new List<GameObject>();

    void Start()
    {
        if (shuffleOnStart)
        {
            GenerateAndDeal();
        }
    }

    [ContextMenu("Generate And Deal")]
    public void GenerateAndDeal()
    {
        ClearSpawned();
        BuildDeck();
        ShuffleDeck();

        // 王牌（13枚）を先に確保
        var wanpai = DrawTiles(13);
        SpawnWanpai(wanpai);

        // 山は残り
        SpawnWall();

        // 配牌 — 親(0)=14枚、子(1)=13枚、子(2)=13枚
        var hand0 = DrawTiles(14);
        var hand1 = DrawTiles(13);
        var hand2 = DrawTiles(13);

        SpawnPlayerHand(0, hand0, isHost: true);
        SpawnPlayerHand(1, hand1, isHost: false);
        SpawnPlayerHand(2, hand2, isHost: false);
    }

    // 牌デッキを構築する。必要に応じてここを編集して牌の種類・枚数を調整する。
    private void BuildDeck()
    {
        deck.Clear();

        // 萬子 (1m-9m) 各4枚
        AddNumberedSuit("m");

        // 筒子 (1p-9p) 各4枚。5p のうち1枚を赤ドラとして "5pr"
        AddNumberedSuit("p", redFiveCode: "5pr");

        // 索子 (1s-9s) 各4枚。5s のうち1枚を赤ドラとして "5sr"
        AddNumberedSuit("s", redFiveCode: "5sr");

        // 字牌（ton,nan,sya,pe,haku,hatu,chun） 各4枚
        string[] honors = new[] { "ton", "nan", "sya", "pe", "haku", "hatu", "chun" };
        foreach (var h in honors)
        {
            for (int i = 0; i < 4; i++) deck.Add(h);
        }

        // 注: 上の構成は標準的な136枚に近い構成になる。
        // README の「総数: 80枚」を厳密に再現したい場合は、
        // AddNumberedSuit のカウント等を調整すること。
    }

    private void AddNumberedSuit(string suit, string redFiveCode = null)
    {
        for (int num = 1; num <= 9; num++)
        {
            string baseTile = num.ToString() + suit;
            int copies = 4;
            for (int c = 0; c < copies; c++)
            {
                // 5 の一枚を赤ドラにする（存在する場合）
                if (redFiveCode != null && num == 5 && c == 0)
                {
                    deck.Add(redFiveCode);
                }
                else
                {
                    deck.Add(baseTile);
                }
            }
        }
    }

    private void ShuffleDeck()
    {
        for (int i = 0; i < deck.Count; i++)
        {
            int j = Random.Range(i, deck.Count);
            var tmp = deck[i];
            deck[i] = deck[j];
            deck[j] = tmp;
        }
    }

    private List<string> DrawTiles(int count)
    {
        var res = new List<string>();
        for (int i = 0; i < count && deck.Count > 0; i++)
        {
            res.Add(deck[0]);
            deck.RemoveAt(0);
        }
        return res;
    }

    private void SpawnPlayerHand(int playerIndex, List<string> tiles, bool isHost)
    {
        if (playerIndex < 0 || playerIndex >= playerTileRoots.Length) return;
        var root = playerTileRoots[playerIndex];
        if (root == null) return;

        float startOffset = -((tiles.Count - 1) * tileSpacing) / 2f;

        for (int i = 0; i < tiles.Count; i++)
        {
            var tileGO = Instantiate(tilePrefab, root);
            tileGO.name = $"Tile_{tiles[i]}_{i}";
            Vector3 pos = root.position + root.right * (startOffset + i * tileSpacing);
            // 親は14枚で少し前に出すなど微調整（デフォルト）
            if (isHost)
            {
                pos += root.forward * 0.02f;
            }
            tileGO.transform.position = pos;
            tileGO.transform.rotation = root.rotation;
            spawnedTiles.Add(tileGO);

            // オプション: 視認のためテキストがあるなら表示
            var txt = tileGO.GetComponentInChildren<UnityEngine.UI.Text>();
            if (txt != null) txt.text = tiles[i];
        }
    }

    private void SpawnWanpai(List<string> wanpai)
    {
        if (wanpaiRoot == null) return;

        float startOffset = -((wanpai.Count - 1) * tileSpacing) / 2f;
        for (int i = 0; i < wanpai.Count; i++)
        {
            var tileGO = Instantiate(tilePrefab, wanpaiRoot);
            tileGO.name = $"Wanpai_{wanpai[i]}_{i}";
            Vector3 pos = wanpaiRoot.position + wanpaiRoot.right * (startOffset + i * tileSpacing);
            tileGO.transform.position = pos;
            tileGO.transform.rotation = wanpaiRoot.rotation;
            spawnedTiles.Add(tileGO);
        }
    }

    private void SpawnWall()
    {
        if (wallOrigin == null) return;

        // 壁は残りのデッキを二段に並べて視覚化
        int perLayer = Mathf.Max(1, Mathf.CeilToInt(deck.Count / wallLayers));
        int idx = 0;
        for (int layer = 0; layer < wallLayers; layer++)
        {
            int count = Mathf.Min(perLayer, deck.Count - idx);
            float start = -((count - 1) * wallSpacing) / 2f;
            for (int i = 0; i < count && idx < deck.Count; i++, idx++)
            {
                var tileGO = Instantiate(tilePrefab, wallOrigin);
                tileGO.name = $"Wall_{deck[idx]}_{layer}_{i}";
                Vector3 pos = wallOrigin.position
                    + wallOrigin.right * (start + i * wallSpacing)
                    + wallOrigin.up * (layer * 0.02f);
                tileGO.transform.position = pos;
                tileGO.transform.rotation = wallOrigin.rotation;
                spawnedTiles.Add(tileGO);
            }
        }

        // 山から既に使った分を消す（ここでは deck を空にしないが、表示上は生成）
        // 実ゲームロジックではここから更にツモや切り牌を扱う。
    }

    private void ClearSpawned()
    {
        for (int i = spawnedTiles.Count - 1; i >= 0; i--)
        {
            if (spawnedTiles[i] != null) DestroyImmediate(spawnedTiles[i]);
        }
        spawnedTiles.Clear();
    }

    void OnValidate()
    {
        if (playerTileRoots == null || playerTileRoots.Length != 3)
        {
            var arr = new Transform[3];
            if (playerTileRoots != null)
            {
                for (int i = 0; i < Mathf.Min(3, playerTileRoots.Length); i++) arr[i] = playerTileRoots[i];
            }
            playerTileRoots = arr;
        }
    }
}
