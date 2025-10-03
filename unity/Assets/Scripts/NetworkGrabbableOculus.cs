using Photon.Pun;
using Photon.Realtime;
using UnityEngine;

[RequireComponent(typeof(PhotonView))]
public class NetworkGrabbableOculus : MonoBehaviourPunCallbacks, IPunOwnershipCallbacks
{
    [Tooltip("Release時にMasterへ所有権を戻す場合はON")]
    [SerializeField]
    public bool returnOwnershipToMasterOnRelease = false;

    public void OnSelected()
    {
        if (photonView != null)
        {
            photonView.RequestOwnership();
        }
    }

    public void OnUnselected()
    {
        if (returnOwnershipToMasterOnRelease && photonView != null)
        {
            photonView.TransferOwnership(PhotonNetwork.MasterClient);
        }
    }

    public void OnOwnershipRequest(PhotonView targetView, Player requestingPlayer)
    {
        if (targetView == photonView)
        {
            photonView.TransferOwnership(requestingPlayer);
        }
    }

    public void OnOwnershipTransfered(PhotonView targetView, Player previousOwner)
    {
        if (targetView == photonView)
        {
            Debug.Log($"Ownership transferred to {targetView.OwnerActorNr} (from {previousOwner?.ActorNumber})");
        }
    }

    public void OnOwnershipTransferFailed(PhotonView targetView, Player senderOfFailedRequest)
    {
        if (targetView == photonView)
        {
            Debug.LogWarning($"Ownership transfer failed. Sender: {senderOfFailedRequest?.ActorNumber}");
        }
    }
}


