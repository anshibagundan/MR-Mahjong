using Photon.Pun;
using Photon.Realtime;
using UnityEngine;
using UnityEngine.EventSystems;

[RequireComponent(typeof(PhotonView))]
public class NetworkOwnershipOnSelect : MonoBehaviourPunCallbacks, IPointerDownHandler, IBeginDragHandler
{
    [Tooltip("選択/ドラッグ時にMasterへ所有権返却するならOFFのまま、必要なら別スクリプトで制御")]
    public bool autoReturnToMasterOnRelease = false;

    void OnMouseDown()
    {
        TryRequestOwnership();
    }

    public void OnPointerDown(PointerEventData eventData)
    {
        TryRequestOwnership();
    }

    public void OnBeginDrag(PointerEventData eventData)
    {
        TryRequestOwnership();
    }

    private void TryRequestOwnership()
    {
        if (photonView != null && !photonView.IsMine)
        {
            photonView.RequestOwnership();
        }
    }
}


