using Photon.Pun;
using Photon.Realtime;
using UnityEngine;
using UnityEngine.EventSystems;

[RequireComponent(typeof(PhotonView))]
public class NetworkOwnershipOnSelect : MonoBehaviourPunCallbacks, IPointerDownHandler, IBeginDragHandler, IPunOwnershipCallbacks
{
    [Tooltip("掴み終了時にMasterへ所有権を返却するか")]
    public bool returnToMasterOnRelease = false;

    [Tooltip("デバッグログを表示するか")]
    public bool enableDebugLogs = true;

    void Awake()
    {
        PhotonNetwork.AddCallbackTarget(this);
    }

    void OnDestroy()
    {
        PhotonNetwork.RemoveCallbackTarget(this);
    }

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

    public void OnGrabStart()
    {
        TryRequestOwnership();
    }

    public void OnGrabEnd()
    {
        if (returnToMasterOnRelease && photonView != null && photonView.IsMine)
        {
            if (enableDebugLogs)
                Debug.Log($"Returning ownership to Master for {gameObject.name}");
            photonView.TransferOwnership(PhotonNetwork.MasterClient);
        }
    }

    private void TryRequestOwnership()
    {
        if (photonView != null && !photonView.IsMine)
        {
            if (enableDebugLogs)
                Debug.Log($"Requesting ownership for {gameObject.name} from {PhotonNetwork.LocalPlayer.ActorNumber}");
            photonView.RequestOwnership();
        }
    }

    public void OnOwnershipRequest(PhotonView targetView, Player requestingPlayer)
    {
        if (targetView == photonView)
        {
            if (enableDebugLogs)
                Debug.Log($"Auto-approving ownership request from {requestingPlayer.ActorNumber}");
            photonView.TransferOwnership(requestingPlayer);
        }
    }

    public void OnOwnershipTransfered(PhotonView targetView, Player previousOwner)
    {
        if (targetView == photonView)
        {
            if (enableDebugLogs)
                Debug.Log($"Ownership transferred to {targetView.OwnerActorNr} (from {previousOwner?.ActorNumber})");
        }
    }

    public void OnOwnershipTransferFailed(PhotonView targetView, Player senderOfFailedRequest)
    {
        if (targetView == photonView)
        {
            if (enableDebugLogs)
                Debug.LogWarning($"Ownership transfer failed. Sender: {senderOfFailedRequest?.ActorNumber}");
        }
    }
}


