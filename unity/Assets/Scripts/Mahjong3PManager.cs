using System.Collections.Generic;
using UnityEngine;
using System.Threading;
using System.Net.WebSockets;
using System;
using System.Linq;
using System.Text;

[System.Serializable]
public class GameStartData {
    public string type;
    public Data data;

    [System.Serializable]
    public class Data {
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

[System.Serializable]
public class TypeOnly
{
    public string type;
}

public class Mahjong3PManager : MonoBehaviour
{
    private ClientWebSocket ws;
    private CancellationTokenSource cts;
    private Dictionary<string, GameObject> tilePrefabs = new Dictionary<string, GameObject>();

    async void Start()
    {
        LoadPrefabs();

        ws = new ClientWebSocket();
        cts = new CancellationTokenSource();

        Uri serverUri = new Uri("ws://localhost:8080/ws/game");

        try
        {
            // WebSocket接続
            await ws.ConnectAsync(serverUri, cts.Token);
            Debug.Log("WebSocket connected!");
        }
        catch (Exception ex)
        {
            Debug.LogError("ConnectAsync failed: " + ex.Message);
            return; // 接続できなければここで終了
        }

        // 接続状態をチェック
        if (ws.State != WebSocketState.Open)
        {
            Debug.LogError("WebSocket is not open after ConnectAsync");
            return;
        }

        try
        {
            // プレイヤーIDを送信
            string sendMessage = "{\"type\":\"connection_check\",\"data\":{\"playerId\":\"p3\"}}";
            var bytesToSend = new ArraySegment<byte>(Encoding.UTF8.GetBytes(sendMessage));
            await ws.SendAsync(bytesToSend, WebSocketMessageType.Text, true, cts.Token);
            Debug.Log("Sent connection_check for p3");
        }
        catch (Exception ex)
        {
            Debug.LogError("SendAsync failed: " + ex.Message);
            return;
        }

        try
        {
            // 受信ループ
            await ReceiveLoop();
        }
        catch (Exception ex)
        {
            Debug.LogError("ReceiveLoop failed: " + ex.Message);
        }
    }

    async System.Threading.Tasks.Task ReceiveLoop()
    {
        var buffer = new byte[1024 * 4];

        while (ws.State == WebSocketState.Open)
        {
            var result = await ws.ReceiveAsync(new ArraySegment<byte>(buffer), cts.Token);
            if (result.MessageType == WebSocketMessageType.Close)
            {
                await ws.CloseAsync(WebSocketCloseStatus.NormalClosure, "", cts.Token);
                Debug.Log("WebSocket closed");
            }
            else
            {
                string msg = Encoding.UTF8.GetString(buffer, 0, result.Count);
                Debug.Log("Received raw JSON: " + msg);

                // まず type だけチェック
                TypeOnly typeCheck = null;
                try
                {
                    typeCheck = JsonUtility.FromJson<TypeOnly>(msg);
                }
                catch (Exception ex)
                {
                    Debug.LogError("Failed to parse type: " + ex.Message);
                    continue;
                }

                if (typeCheck == null || string.IsNullOrEmpty(typeCheck.type))
                {
                    Debug.LogWarning("Received message without type, skipping");
                    continue;
                }

                if (typeCheck.type == "game_start")
                {
                    // game_start の場合のみ GameStartData にパース
                    GameStartData gameData = null;
                    try
                    {
                        gameData = JsonUtility.FromJson<GameStartData>(msg);
                        Debug.Log("wanpai JSON: " + JsonUtility.ToJson(gameData.data.wanpai, true));
                        Debug.Log($"yama count: {gameData?.data?.yama?.Count}");
                        Debug.Log($"players count: {gameData?.data?.players?.Count}");
                    }
                    catch (Exception ex)
                    {
                        Debug.LogError("GameStartData deserialization failed: " + ex.Message);
                        continue;
                    }

                    if (gameData?.data != null)
                    {
                        SetupGame(gameData);
                    }
                    else
                    {
                        Debug.LogError("gameData.data is null!");
                    }
                }
                else
                {
                    // それ以外の type の場合はログだけ出して無視
                    Debug.Log($"Non-game-start message received, type: {typeCheck.type}");
                }
            }
        }
    }

    void SetupGame(GameStartData gameData)
    {
        // 山を配置（南・東・西）
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
