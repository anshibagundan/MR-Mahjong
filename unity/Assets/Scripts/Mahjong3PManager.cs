using System.Collections.Generic;
using UnityEngine;
using System;
using System.Linq;
using System.Text;
using Photon.Pun;
using Photon.Realtime;
using WebSocketSharp;
using System.Collections;

[System.Serializable]
public class ConnectionResponse
{
    public string type;
    public ConnectionData data;

    [System.Serializable]
    public class ConnectionData
    {
        public string playerId;
        public int playersCount;
        public int maxPlayers;
        public string status;
        public string message;
    }
}

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

public class Mahjong3PManager : MonoBehaviourPunCallbacks
{
    public GameObject PhotonFailureObject;
    public string RoomName = "MahjongRoom";
    public int MaxPlayers = 3;
    public float ConnectionTimeout = 10f;
    public int MaxRetryAttempts = 3;
    public bool isFirstPlayer = false;

    private WebSocket ws;
    private Dictionary<string, GameObject> tilePrefabs = new Dictionary<string, GameObject>();
    private int retryCount = 0;
    private bool isConnecting = false;
    private int webSocketRetryCount = 0;
    private readonly Queue<Action> _mainThreadActions = new Queue<Action>();
    private readonly object _actionsLock = new object();

    void Start()
    {
        LoadPrefabs();

        // Photonサーバーに接続
        if (!PhotonNetwork.IsConnected)
        {
            PhotonNetwork.ConnectUsingSettings();
            Debug.Log("Connecting to Photon...");
        }
    }

    void Update()
    {
        // 実体生成などはメインスレッドで実行
        if (_mainThreadActions.Count == 0) return;
        lock (_actionsLock)
        {
            while (_mainThreadActions.Count > 0)
            {
                var action = _mainThreadActions.Dequeue();
                try { action?.Invoke(); }
                catch (Exception ex) { Debug.LogError($"MainThread action error: {ex.Message}"); }
            }
        }
    }


    void OnApplicationPause(bool pauseStatus)
    {
        if (pauseStatus)
        {
            Debug.Log("Application paused - maintaining connection");
        }
        else
        {
            Debug.Log("Application resumed - checking connection");
            if (!PhotonNetwork.IsConnected && !isConnecting)
            {
                Debug.Log("Reconnecting to Photon after resume");
                isConnecting = true;
                PhotonNetwork.ConnectUsingSettings();
            }
        }
    }

    void OnApplicationFocus(bool hasFocus)
    {
        if (hasFocus)
        {
            Debug.Log("Application gained focus - checking connections");
            if (!PhotonNetwork.IsConnected && !isConnecting)
            {
                Debug.Log("Reconnecting to Photon after focus");
                isConnecting = true;
                PhotonNetwork.ConnectUsingSettings();
            }
        }
        else
        {
            Debug.Log("Application lost focus");
        }
    }

    // Photon マスターサーバーに接続成功時
    public override void OnConnectedToMaster()
    {
        Debug.Log("Connected to Photon Master Server");
        isConnecting = false;
        retryCount = 0;
        // 固定ルームに参加
        PhotonNetwork.JoinRoom(RoomName);
    }

    // ルーム参加失敗時（ルームが存在しない場合）
    public override void OnJoinRoomFailed(short returnCode, string message)
    {
        Debug.Log($"Failed to join room: {message}. Creating new room...");
        // ルームを新規作成（最大3人まで）
        RoomOptions roomOptions = new RoomOptions();
        roomOptions.MaxPlayers = (byte)MaxPlayers;
        PhotonNetwork.CreateRoom(RoomName, roomOptions);
    }

    // ルーム参加成功時
    public override void OnJoinedRoom()
    {
        Debug.Log($"Joined room: {PhotonNetwork.CurrentRoom.Name}");
        Debug.Log($"Players in room: {PhotonNetwork.CurrentRoom.PlayerCount}/{PhotonNetwork.CurrentRoom.MaxPlayers}");

        // WebSocket接続を開始
        StartWebSocketConnection();
    }

    // Photon切断時
    public override void OnDisconnected(DisconnectCause cause)
    {
        base.OnDisconnected(cause);
        isConnecting = false;

        Debug.LogError("Disconnected from Photon: " + cause.ToString());

        // タイムアウトや一時的な接続問題の場合は再試行
        if ((cause == DisconnectCause.ClientTimeout || cause == DisconnectCause.ServerTimeout ||
             cause == DisconnectCause.DisconnectByServerLogic || cause == DisconnectCause.DisconnectByServerReasonUnknown ||
             cause == DisconnectCause.Exception || cause == DisconnectCause.ExceptionOnConnect) &&
            retryCount < MaxRetryAttempts)
        {
            retryCount++;
            Debug.Log($"Attempting reconnection {retryCount}/{MaxRetryAttempts}...");
            Invoke(nameof(RetryConnection), 2f); // 2秒後に再試行
            return;
        }

        // 最終的に接続に失敗した場合、PhotonFailureObjectを表示
        if (PhotonFailureObject != null)
        {
            // ローカルにオブジェクトをInstantiateする例
            Instantiate(PhotonFailureObject, new Vector3(0f, 0f, 0f), Quaternion.identity);
        }
        else
        {
            Debug.LogWarning("PhotonFailureObject is not set in the inspector.");
        }
    }

    void RetryConnection()
    {
        if (!PhotonNetwork.IsConnected && !isConnecting)
        {
            isConnecting = true;
            PhotonNetwork.ConnectUsingSettings();
        }
    }

    void RetryWebSocketConnection()
    {
        if (ws?.ReadyState != WebSocketState.Open)
        {
            Debug.Log("Retrying WebSocket connection...");
            StartWebSocketConnection();
        }
    }

    // WebSocket接続処理（Photon接続後に実行）
    void StartWebSocketConnection()
    {
        if (ws != null && ws.ReadyState == WebSocketState.Open)
        {
            Debug.Log("WebSocket already connected");
            return;
        }

        Debug.Log("Starting WebSocket connection...");

        string serverUrl = "wss://app-37aa2340-1e0d-4ba5-aedf-d0383cb98c14.ingress.apprun.sakura.ne.jp/ws/game";
        ws = new WebSocket(serverUrl);

        // WebSocketイベントハンドラーを設定
        ws.OnOpen += (sender, e) => {
            Debug.Log("WebSocket connected successfully!");
            webSocketRetryCount = 0; // 接続成功時にリトライカウンターをリセット

            // プレイヤーIDを送信（UUIDを使用）
            string playerId = System.Guid.NewGuid().ToString();
            string sendMessage = $"{{\"type\":\"connection_check\",\"data\":{{\"playerId\":\"{playerId}\"}}}}";
            Debug.Log($"Sending message: {sendMessage}");
            ws.Send(sendMessage);
            Debug.Log($"Sent connection_check for {playerId}");
        };

        ws.OnMessage += (sender, e) => {
            // WebSocketSharpのコールバックは別スレッドの場合があるため、メインスレッドにディスパッチ
            var payload = e.Data;
            lock (_actionsLock)
            {
                _mainThreadActions.Enqueue(() => {
                    ProcessWebSocketMessage(payload);
                });
            }
        };

        ws.OnError += (sender, e) => {
            Debug.LogError($"WebSocket error: {e.Message}");
        };

        ws.OnClose += (sender, e) => {
            Debug.LogWarning($"WebSocket closed: {e.Code}, {e.Reason}");

            // WebSocket接続に失敗した場合の再試行
            if (webSocketRetryCount < MaxRetryAttempts)
            {
                webSocketRetryCount++;
                Debug.Log($"WebSocket connection failed. Retrying {webSocketRetryCount}/{MaxRetryAttempts} in 3 seconds...");
                Invoke(nameof(RetryWebSocketConnection), 3f);
            }
            else if (webSocketRetryCount >= MaxRetryAttempts)
            {
                Debug.LogError("WebSocket connection failed after maximum retry attempts");
            }
        };

        Debug.Log($"Connecting to: {serverUrl}");
        ws.Connect();
    }

    void ProcessWebSocketMessage(string message)
    {
        // まず type だけチェック
        TypeOnly typeCheck = null;
        try
        {
            typeCheck = JsonUtility.FromJson<TypeOnly>(message);
        }
        catch (Exception ex)
        {
            Debug.LogError("Failed to parse type: " + ex.Message);
            return;
        }

        if (typeCheck == null || string.IsNullOrEmpty(typeCheck.type))
        {
            Debug.LogWarning("Received message without type, skipping");
            return;
        }

        if (typeCheck.type == "game_start")
        {
            // game_start の場合のみ GameStartData にパース
            GameStartData gameData = null;
            try
            {
                gameData = JsonUtility.FromJson<GameStartData>(message);
                Debug.Log($"Game start - yama: {gameData?.data?.yama?.Count} tiles, players: {gameData?.data?.players?.Count}");
            }
            catch (Exception ex)
            {
                Debug.LogError("GameStartData deserialization failed: " + ex.Message);
                return;
            }

            if (gameData?.data != null)
            {
                if (isFirstPlayer)
                {
                    Debug.Log("First player - instantiating objects");
                    SetupGame(gameData);
                }
                else
                {
                    Debug.Log("Not first player - do not instantiate yet");
                }
            }
            else
            {
                Debug.LogError("gameData.data is null!");
            }
        }
        else if (typeCheck.type == "connection_response")
        {
            ConnectionResponse connResp = JsonUtility.FromJson<ConnectionResponse>(message);
            if (connResp?.data != null)
            {
                isFirstPlayer = connResp.data.playersCount == 1;
                Debug.Log($"isFirstPlayer set to {isFirstPlayer}");
            }
        }
    }

    void OnDestroy()
    {
        // WebSocket接続をクリーンアップ
        if (ws != null)
        {
            try
            {
                if (ws.ReadyState == WebSocketState.Open)
                {
                    ws.Close();
                }
                ws = null;
            }
            catch (Exception ex)
            {
                Debug.LogError($"Error closing WebSocket: {ex.Message}");
            }
        }
    }

    void SetupGame(GameStartData gameData)
    {
        Debug.Log("Setting up game board...");

        // 山を配置（南・東・西）
        if (gameData.data.yama.Count > 0)
        {
            PlaceMountain(gameData.data.yama.GetRange(0, Mathf.Min(36, gameData.data.yama.Count)),
                        new Vector3(0, 1.221f, -0.4f), Quaternion.identity); // 南
        }

        if (gameData.data.yama.Count > 36)
        {
            PlaceMountain(gameData.data.yama.GetRange(36, Mathf.Min(36, gameData.data.yama.Count - 36)),
                        new Vector3(0.4f, 1.221f, 0), Quaternion.Euler(0, 90, 0)); // 東
        }

        if (gameData.data.yama.Count > 72)
        {
            PlaceMountain(gameData.data.yama.GetRange(72, Mathf.Min(36, gameData.data.yama.Count - 72)),
                        new Vector3(-0.4f, 1.221f, 0), Quaternion.Euler(0, -90, 0)); // 西
        }

        // ワンパイを配置
        PlaceWanpais(gameData.data.wanpai, new Vector3(0, 1.221f, 0.4f), Quaternion.Euler(0, 180, 0));

        // 手牌を配置
        PlaceHand(gameData.data.players[0].tehai, new Vector3(0, 1.221f, -0.7f), Quaternion.identity);
        PlaceHand(gameData.data.players[1].tehai, new Vector3(0.7f, 1.221f, 0), Quaternion.Euler(0, 270, 0));
        PlaceHand(gameData.data.players[2].tehai, new Vector3(0, 1.221f, 0.7f), Quaternion.Euler(0, 180, 0));

        Debug.Log("Game board setup completed");
    }

    void LoadPrefabs()
    {
        // JSONの牌名と実際のプレハブファイル名の対応
        string[] tileNames = {
            "1p","2p","3p","4p","5p","6p","7p","8p","9p","5pr",
            "1m","2m","3m","4m","5mr","6m","7m","8m","9m",  // 5mは5mrに修正（実際のプレハブファイル名）
            "1s","2s","3s","4s","5s","6s","7s","8s","9s","5sr",
            "ton","nan","sya","pe","haku","hatu","chun"
        };

        int loadedCount = 0;
        foreach (var name in tileNames)
        {
            var prefab = Resources.Load<GameObject>($"Prefabs/{name}");
            if (prefab != null)
            {
                tilePrefabs[name] = prefab;
                loadedCount++;
            }
            else
            {
                Debug.LogWarning($"Prefab not found: {name}");
            }
        }
        Debug.Log($"Loaded {loadedCount}/{tileNames.Length} tile prefabs");
    }

    // JSONの牌名を実際のプレハブ名にマッピング
    string MapTileName(string jsonTileName)
    {
        // 特別な牌名のマッピング
        switch (jsonTileName)
        {
            case "5pr": return "5pr";  // 赤5p
            case "5sr": return "5sr";  // 赤5s
            default: return jsonTileName;  // その他はそのまま
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
        int instantiatedCount = 0;

        for (int i = 0; i < tilesPerMountain; i++)
        {
            string jsonTileName = tiles[i];
            string tileName = MapTileName(jsonTileName);  // 牌名をマッピング

            if (!tilePrefabs.ContainsKey(tileName))
            {
                Debug.LogWarning($"Tile prefab not found: {tileName} (from JSON: {jsonTileName})");
                continue;
            }

            int row = i % 2;  // 0 or 1
            int col = i / 2;  // 0〜17

            Vector3 localPos = new Vector3((8.5f - col) * tileWidth, 0.0041f + (row * tileHeight), 0f);
            Quaternion localRot = Quaternion.identity;

            // 親のTransform基準でワールド座標と回転に変換
            Vector3 worldPos = parent.transform.TransformPoint(localPos);
            Quaternion worldRot = parent.transform.rotation * localRot;

            try
            {
                GameObject tile = PhotonNetwork.Instantiate($"Prefabs/{tileName}", worldPos, worldRot);
                tile.transform.SetParent(parent.transform, true); // 親子付け
                instantiatedCount++;
            }
            catch (Exception ex)
            {
                Debug.LogError($"Failed to instantiate tile {tileName}: {ex.Message}");
            }
        }

        Debug.Log($"Mountain placed: {instantiatedCount}/{tilesPerMountain} tiles");
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

            Vector3 pos = parent.transform.position + new Vector3(col * spacingX, row * spacingY, 0);

            GameObject tile = PhotonNetwork.Instantiate(
                $"Prefabs/{tileName}",  // Resources/Prefabs/ にあるプレハブ名
                pos,                   // ワールド座標
                parent.transform.rotation, // 親の回転を継承
                0
            );

            // Photon は親子関係を自動では同期しないので、自分で設定する
            tile.transform.SetParent(parent.transform);
            tile.transform.localPosition = new Vector3(col * spacingX, row * spacingY, 0);
        }

        // ===== Revealed Dora =====
        for (int i = 0; i < wanpai.revealedDora.Count; i++)
        {
            string tileName = wanpai.revealedDora[i];
            if (!tilePrefabs.ContainsKey(tileName)) continue;

            int row = 1;             // 2行目
            int col = 2 + i;         // 2列目以降
            // ローカル座標での位置・回転
            Vector3 localPos = new Vector3(col * spacingX, row * spacingY, 0);
            Quaternion localRot = Quaternion.Euler(0, 0, 180);

            // 親のTransformを基準にワールド座標へ変換
            Vector3 worldPos = parent.transform.TransformPoint(localPos);
            Quaternion worldRot = parent.transform.rotation * localRot;

            // Photonで生成
            GameObject dora = PhotonNetwork.Instantiate($"Prefabs/{tileName}", worldPos, worldRot);

            // 親子付け（Photonではparentを直接渡せないので後付け）
            dora.transform.SetParent(parent.transform, true);
        }

        // ===== Unrevealed Dora =====
        for (int i = 0; i < wanpai.unrevealedDoras.Count; i++)
        {
            string tileName = wanpai.unrevealedDoras[i];
            if (!tilePrefabs.ContainsKey(tileName)) continue;

            int row = 0;             // 2行目
            int col = 6 - i;             // 左から順
            Vector3 pos = parent.transform.position + new Vector3(col * spacingX, row * spacingY, 0);

            GameObject tile = PhotonNetwork.Instantiate(
                $"Prefabs/{tileName}",  // Resources/Prefabs/ にあるプレハブ名
                pos,                   // ワールド座標
                parent.transform.rotation, // 親の回転を継承
                0
            );

            // Photon は親子関係を自動では同期しないので、自分で設定する
            tile.transform.SetParent(parent.transform);
            tile.transform.localPosition = new Vector3(col * spacingX, row * spacingY, 0);
        }

        // ===== Kan Doras =====
        for (int i = 0; i < wanpai.kanDoras.Count; i++)
        {
            string tileName = wanpai.kanDoras[i];
            if (!tilePrefabs.ContainsKey(tileName)) continue;

            int row = 1;             // 3行目
            int col = 6 - i;             // 左から順
            Vector3 pos = parent.transform.position + new Vector3(col * spacingX, row * spacingY, 0);

            GameObject tile = PhotonNetwork.Instantiate(
                $"Prefabs/{tileName}",  // Resources/Prefabs/ にあるプレハブ名
                pos,                   // ワールド座標
                parent.transform.rotation, // 親の回転を継承
                0
            );

            // Photon は親子関係を自動では同期しないので、自分で設定する
            tile.transform.SetParent(parent.transform);
            tile.transform.localPosition = new Vector3(col * spacingX, row * spacingY, 0);
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
            string jsonTileName = handTiles[i];
            string tileName = MapTileName(jsonTileName);  // 牌名をマッピング

            if (!tilePrefabs.ContainsKey(tileName))
            {
                Debug.LogWarning($"Hand tile prefab not found: {tileName} (from JSON: {jsonTileName})");
                continue;
            }

            Vector3 localPos = new Vector3(offset + i * spacing, 0.0041f, 0);
            Quaternion localRot = Quaternion.Euler(90, 0, 0);

            // 親（親Transformのワールド座標に変換する）
            Vector3 worldPos = parent.transform.TransformPoint(localPos);
            Quaternion worldRot = parent.transform.rotation * localRot;

            // Photonで生成（ワールド座標とワールド回転を渡す）
            GameObject tehai = PhotonNetwork.Instantiate($"Prefabs/{tileName}", worldPos, worldRot);

            // 親子付け（PhotonNetwork.Instantiate では parent 指定できないので後付けする必要あり）
            tehai.transform.SetParent(parent.transform, true);
        }
    }
}
