using System.Collections.Generic;
using UnityEngine;
using System.Threading;
using System.Net.WebSockets;
using System;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Photon.Pun;
using Photon.Realtime;

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

    private ClientWebSocket ws;
    private CancellationTokenSource cts;
    private Dictionary<string, GameObject> tilePrefabs = new Dictionary<string, GameObject>();
    private int retryCount = 0;
    private bool isConnecting = false;

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
        _ = StartWebSocketConnection();
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

    // WebSocket接続処理（Photon接続後に実行）
    async Task StartWebSocketConnection()
    {
        if (ws != null && ws.State == WebSocketState.Open)
        {
            Debug.Log("WebSocket already connected");
            return;
        }

        ws = new ClientWebSocket();
        cts = new CancellationTokenSource();
        ws.Options.RemoteCertificateValidationCallback = (sender, cert, chain, errors) => true;

        Uri serverUri = new Uri("wss://app-37aa2340-1e0d-4ba5-aedf-d0383cb98c14.ingress.apprun.sakura.ne.jp/ws/game");

        try
        {
            // タイムアウト設定
            using (var timeoutCts = new CancellationTokenSource(TimeSpan.FromSeconds(ConnectionTimeout)))
            {
                await ws.ConnectAsync(serverUri, timeoutCts.Token);
            }

            Debug.Log("WebSocket connected!");

            if (ws.State != WebSocketState.Open)
            {
                Debug.LogError("WebSocket is not open after ConnectAsync");
                return;
            }

            // プレイヤーIDを送信（UUIDを使用）
            string playerId = System.Guid.NewGuid().ToString();
            string sendMessage = $"{{\"type\":\"connection_check\",\"data\":{{\"playerId\":\"{playerId}\"}}}}";
            var bytesToSend = new ArraySegment<byte>(Encoding.UTF8.GetBytes(sendMessage));
            await ws.SendAsync(bytesToSend, WebSocketMessageType.Text, true, cts.Token);
            Debug.Log($"Sent connection_check for {playerId}");

            // 受信ループ
            await ReceiveLoop();
        }
        catch (OperationCanceledException)
        {
            Debug.LogError("WebSocket connection timeout");
        }
        catch (System.Net.WebSockets.WebSocketException wsEx)
        {
            Debug.LogError($"WebSocket error: {wsEx.Message}");
        }
        catch (System.Net.Http.HttpRequestException httpEx)
        {
            Debug.LogError($"HTTP error: {httpEx.Message}");
        }
        catch (Exception ex)
        {
            Debug.LogError($"ConnectAsync failed: {ex.Message}");
        }
    }

    async Task ReceiveLoop()
    {
        var buffer = new byte[1024 * 4];

        try
        {
            while (ws != null && ws.State == WebSocketState.Open)
            {
                var result = await ws.ReceiveAsync(new ArraySegment<byte>(buffer), cts.Token);
                if (result.MessageType == WebSocketMessageType.Close)
                {
                    await ws.CloseAsync(WebSocketCloseStatus.NormalClosure, "", cts.Token);
                    Debug.Log("WebSocket closed");
                    break;
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
        catch (OperationCanceledException)
        {
            Debug.Log("WebSocket receive cancelled");
        }
        catch (Exception ex)
        {
            Debug.LogError($"WebSocket receive error: {ex.Message}");
        }
    }

    void OnDestroy()
    {
        // WebSocket接続をクリーンアップ
        if (cts != null)
        {
            cts.Cancel();
            cts.Dispose();
        }

        if (ws != null && ws.State == WebSocketState.Open)
        {
            try
            {
                ws.CloseAsync(WebSocketCloseStatus.NormalClosure, "Application closing", CancellationToken.None);
            }
            catch (Exception ex)
            {
                Debug.LogError($"Error closing WebSocket: {ex.Message}");
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

            GameObject tile = PhotonNetwork.Instantiate(
                $"Prefabs/{tileName}",              // プレハブ名（Resources 下にあること）
                parent.transform.position + new Vector3(x, y, z), // ワールド座標に変換
                parent.transform.rotation,          // 親の回転を継承
                0                                   // group
            );
            tile.transform.localPosition = new Vector3(x, y, z);
            tile.transform.localRotation = Quaternion.identity;
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
            string tileName = handTiles[i];
            if (!tilePrefabs.ContainsKey(tileName)) continue;

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
