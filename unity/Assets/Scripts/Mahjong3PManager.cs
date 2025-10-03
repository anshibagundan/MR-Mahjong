using System.Collections.Generic;
using UnityEngine;

[System.Serializable]
public class GameStartData {
    public string type;
    public Data data;

    [System.Serializable]
    public class Data {
        public string gameId;
        public string playerId;
        public List<string> tehai;
        public Wanpai wanpai;
        public List<string> yama;
        public List<Player> players;
    }

    [System.Serializable]
    public class Wanpai {
        public List<string> revealedDora;
        public List<string> unrevealedDoras;
        public List<string> kanDoras;
        public List<string> rinsyan;
    }

    [System.Serializable]
    public class Player {
        public string id;
        public List<string> tehai;
        public bool isHost;
    }
}

public class Mahjong3PManager : MonoBehaviour
{
    private Dictionary<string, GameObject> tilePrefabs = new Dictionary<string, GameObject>();

    void Start()
    {
        LoadPrefabs();

        // ==== JSON相当のデータをハードコード ====
        string jsonString = @"
        {
            ""type"": ""game_start"",
            ""data"": {
                ""playerId"": ""p1"",
                ""tehai"": [
                ""nan"",
                ""chun"",
                ""chun"",
                ""hatu"",
                ""6p"",
                ""4p"",
                ""4s"",
                ""6p"",
                ""9p"",
                ""6s"",
                ""hatu"",
                ""7p"",
                ""4p"",
                ""2s""
                ],
                ""wanpai"": {
                ""revealedDora"": [
                    ""8s""
                ],
                ""kanDoras"": [
                    ""sya"",
                    ""2p"",
                    ""9p"",
                    ""7s""
                ],
                ""unrevealedDoras"": [
                    ""8s"",
                    ""4s"",
                    ""2s"",
                    ""9s"",
                    ""ton""
                ],
                ""rinsyan"": [
                    ""5p"",
                    ""3p"",
                    ""5s"",
                    ""5s""
                ]
                },
                ""yama"": [
                ""3p"",
                ""3s"",
                ""4s"",
                ""7p"",
                ""chun"",
                ""sya"",
                ""2p"",
                ""8p"",
                ""9m"",
                ""2s"",
                ""nan"",
                ""3s"",
                ""2p"",
                ""6s"",
                ""sya"",
                ""haku"",
                ""9p"",
                ""5p"",
                ""nan"",
                ""ton"",
                ""7s"",
                ""nan"",
                ""1m"",
                ""5pr"",
                ""4p"",
                ""hatu"",
                ""9m"",
                ""3p"",
                ""pe"",
                ""haku"",
                ""5s"",
                ""7p"",
                ""1m"",
                ""9p"",
                ""1p"",
                ""9s"",
                ""7s"",
                ""4s"",
                ""5sr"",
                ""8p"",
                ""2s"",
                ""8p"",
                ""9m"",
                ""7s"",
                ""1p"",
                ""8s"",
                ""9s"",
                ""1s"",
                ""4p"",
                ""pe"",
                ""pe"",
                ""1s"",
                ""2p"",
                ""chun""
                ],
                ""players"": [
                {
                    ""id"": ""p1"",
                    ""tehai"": [
                    ""nan"",
                    ""chun"",
                    ""chun"",
                    ""hatu"",
                    ""6p"",
                    ""4p"",
                    ""4s"",
                    ""6p"",
                    ""9p"",
                    ""6s"",
                    ""hatu"",
                    ""7p"",
                    ""4p"",
                    ""2s""
                    ],
                    ""isHost"": true
                },
                {
                    ""id"": ""p2"",
                    ""tehai"": [
                    ""5p"",
                    ""ton"",
                    ""7p"",
                    ""3p"",
                    ""1m"",
                    ""6s"",
                    ""3s"",
                    ""8s"",
                    ""1s"",
                    ""6p"",
                    ""8p"",
                    ""1m"",
                    ""1p""
                    ],
                    ""isHost"": false
                },
                {
                    ""id"": ""p3"",
                    ""tehai"": [
                    ""1s"",
                    ""haku"",
                    ""6p"",
                    ""sya"",
                    ""haku"",
                    ""3s"",
                    ""pe"",
                    ""hatu"",
                    ""1p"",
                    ""6s"",
                    ""ton"",
                    ""9s"",
                    ""9m""
                    ],
                    ""isHost"": false
                }
                ]
            }
        }";

        // JSON → GameStartDataに変換
        GameStartData gameData = JsonUtility.FromJson<GameStartData>(jsonString);
        // ===============================

        // 山を配置（南・東・西）※yamaを分割して正確な枚数を生成
        if (gameData.data.yama.Count > 0)
            PlaceMountain(gameData.data.yama.GetRange(0, Mathf.Min(36, gameData.data.yama.Count)),
                        new Vector3(0, 1.221f, -0.4f), Quaternion.identity); // 南

        if (gameData.data.yama.Count > 36)
            PlaceMountain(gameData.data.yama.GetRange(36, Mathf.Min(36, gameData.data.yama.Count - 36)),
                        new Vector3(0.4f, 1.221f, 0), Quaternion.Euler(0, 90, 0)); // 東

        if (gameData.data.yama.Count > 72)
            PlaceMountain(gameData.data.yama.GetRange(72, Mathf.Min(36, gameData.data.yama.Count - 72)),
                        new Vector3(-0.4f, 1.221f, 0), Quaternion.Euler(0, -90, 0)); // 西

        // ワンパイを配置
        PlaceWanpais(gameData.data.wanpai, new Vector3(0, 1.221f, 0.4f), Quaternion.Euler(0, 180, 0));

        // 手牌を配置
        PlaceHand(gameData.data.players[0].tehai, new Vector3(0, 1.221f, -0.7f), Quaternion.identity);
        PlaceHand(gameData.data.players[1].tehai, new Vector3(0.7f, 1.221f, 0), Quaternion.Euler(0, 270, 0));
        PlaceHand(gameData.data.players[2].tehai, new Vector3(0, 1.221f, 0.7f), Quaternion.Euler(0, 180, 0));
    }

    void LoadPrefabs()
    {
        string[] tileNames = {
            "1p","2p","3p","4p","5p","6p","7p","8p","9p","5pr",
            "1m","2m","3m","4m","5m","6m","7m","8m","9m",
            "1s","2s","3s","4s","5s","6s","7s","8s","9s","5sr",
            "ton","nan","sya","pe","haku","hatu","chun"
        };

        foreach (var name in tileNames)
        {
            var prefab = Resources.Load<GameObject>($"Prefabs/{name}");
            if (prefab != null)
            {
                tilePrefabs[name] = prefab;
            }
            else
            {
                Debug.LogWarning($"Prefab not found: {name}");
            }
        }
    }

    void PlaceMountain(List<string> tiles, Vector3 centerPos, Quaternion rotation)
    {
        float tileWidth = 0.041f;
        float tileHeight = 0.032f;

        GameObject parent = new GameObject("Mountain");
        parent.transform.position = centerPos;
        parent.transform.rotation = rotation;

        int tilesPerMountain = Mathf.Min(tiles.Count, 36); // 山1つ分

        for (int i = 0; i < tilesPerMountain; i++)
        {
            string tileName = tiles[i];
            if (!tilePrefabs.ContainsKey(tileName)) continue;

            int row = i % 2;  // 0 or 1
            int col = i / 2;  // 0〜17

            float x = (8.5f - col) * tileWidth; // 中心揃え
            float y = row * tileHeight - 0.00739f;
            float z = 0f; // 山の奥行き方向は固定（前後に重ねない）

            Instantiate(tilePrefabs[tileName], parent.transform)
                .transform.localPosition = new Vector3(x, y, z);
        }
    }

    void PlaceWanpais(GameStartData.Wanpai wanpai, Vector3 centerPos, Quaternion rotation)
    {
        float spacingX = 0.041f; // 左右間隔
        float spacingY = 0.032f;  // 高さ間隔

        GameObject parent = new GameObject("Wanpai");
        parent.transform.position = centerPos;
        parent.transform.rotation = rotation; // ここで回転も使う

        // ===== Rinsyan =====
        for (int i = 0; i < wanpai.rinsyan.Count; i++)
        {
            string tileName = wanpai.rinsyan[i];
            if (!tilePrefabs.ContainsKey(tileName)) continue;

            int row = i % 2;
            int col = i / 2;

            Vector3 pos = new Vector3(col * spacingX, row * spacingY, 0);
            Instantiate(tilePrefabs[tileName], parent.transform).transform.localPosition = pos;
        }

        // ===== Revealed Dora =====
        for (int i = 0; i < wanpai.revealedDora.Count; i++)
        {
            string tileName = wanpai.revealedDora[i];
            if (!tilePrefabs.ContainsKey(tileName)) continue;

            int row = 1;             // 2行目
            int col = 2 + i;         // 2列目以降
            Vector3 pos = new Vector3(col * spacingX, row * spacingY, 0);
            GameObject dora = Instantiate(tilePrefabs[tileName], parent.transform);
            dora.transform.localPosition = pos;
            dora.transform.localRotation = Quaternion.Euler(0, 0, 180); // z軸180°回転
        }

        // ===== Unrevealed Dora =====
        for (int i = 0; i < wanpai.unrevealedDoras.Count; i++)
        {
            string tileName = wanpai.unrevealedDoras[i];
            if (!tilePrefabs.ContainsKey(tileName)) continue;

            int row = 0;             // 2行目
            int col = 6 - i;             // 左から順
            Vector3 pos = new Vector3(col * spacingX, row * spacingY, 0);
            Instantiate(tilePrefabs[tileName], parent.transform).transform.localPosition = pos;
        }

        // ===== Kan Doras =====
        for (int i = 0; i < wanpai.kanDoras.Count; i++)
        {
            string tileName = wanpai.kanDoras[i];
            if (!tilePrefabs.ContainsKey(tileName)) continue;

            int row = 1;             // 3行目
            int col = 6 - i;             // 左から順
            Vector3 pos = new Vector3(col * spacingX, row * spacingY, 0);
            Instantiate(tilePrefabs[tileName], parent.transform).transform.localPosition = pos;
        }
    }



    void PlaceHand(List<string> handTiles, Vector3 centerPos, Quaternion rotation, bool isSelf = false)
    {
        float spacing = 0.041f;
        GameObject parent = new GameObject("Hand");
        parent.transform.position = centerPos;
        parent.transform.rotation = rotation;

        int count = handTiles.Count;
        float offset = -(count - 1) * spacing / 2f;

        for (int i = 0; i < count; i++)
        {
            string tileName = handTiles[i];
            if (!tilePrefabs.ContainsKey(tileName)) continue;

            Vector3 pos = new Vector3(offset + i * spacing, 0.0041f, 0);
            GameObject tehai = Instantiate(tilePrefabs[tileName], parent.transform);
            tehai.transform.localPosition = pos;
            tehai.transform.localRotation = Quaternion.Euler(90, 0, 0);
        }
    }
}
