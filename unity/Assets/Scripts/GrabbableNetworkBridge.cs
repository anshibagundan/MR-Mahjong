using UnityEngine;
using Oculus.Interaction;

public class GrabbableNetworkBridge : MonoBehaviour
{
    private Grabbable grabbable;
    private NetworkOwnershipOnSelect networkOwnership;
    private int lastGrabPointsCount = 0;

    void Start()
    {
        grabbable = GetComponent<Grabbable>();
        networkOwnership = GetComponent<NetworkOwnershipOnSelect>();
    }

    void Update()
    {
        if (grabbable == null || networkOwnership == null) return;

        // GrabPoints の数で掴み状態を判定
        int currentGrabPointsCount = grabbable.GrabPoints.Count;

        // 掴み開始を検出（0 → 1以上）
        if (currentGrabPointsCount > 0 && lastGrabPointsCount == 0)
        {
            networkOwnership.OnGrabStart();
        }

        // 掴み終了を検出（1以上 → 0）
        if (currentGrabPointsCount == 0 && lastGrabPointsCount > 0)
        {
            networkOwnership.OnGrabEnd();
        }

        lastGrabPointsCount = currentGrabPointsCount;
    }
}
