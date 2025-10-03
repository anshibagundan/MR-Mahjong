using UnityEngine;
using UnityEngine.Android;

public class MicrophonePermissionHelper : MonoBehaviour
{
    void Start()
    {
        // アプリがマイクの使用権限を持っているかチェック
        if (!Permission.HasUserAuthorizedPermission(Permission.Microphone))
        {
            // 権限がない場合、ユーザーに許可を求めるダイアログを表示
            Permission.RequestUserPermission(Permission.Microphone);
        }
    }
}
