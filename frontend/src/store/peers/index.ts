import { createStore, createEffect, createEvent } from 'effector';

import { fetchData } from '../utils';
import {
  PeerType,
  PeerStoreType,
  PeerInfoType,
  PeerRecordType,
  FlatPeerType,
  PeerSetEditingType,
  PeersWireguard
} from './types';
import { GET_IPV4_URL, PEERS_URL, PEERS_WIREGUARD_URL } from './constants';

const initialPeersStore: PeerStoreType = {
  peers: [],
  peerToSave: null
};

export const $peersStore = createStore(initialPeersStore);

export const setPeers = createEvent<PeerRecordType[]>();
export const createPeer = createEvent<string>();
export const cancelCreatePeer = createEvent();
export const savePeer = createEvent<PeerInfoType>();
export const deletePeer = createEvent<number>();
export const changePeer = createEvent<PeerInfoType>();
export const setIsEditing = createEvent<PeerSetEditingType>();

export const getAllPeersFx = createEffect<void, PeerRecordType[], Response>(
  () => fetchData(PEERS_URL).then((res) => res.json())
);

export const getPeersWireguardFx = createEffect<{private_key: string; ipv4: string}, PeersWireguard, Response>(
  (data) => fetchData(PEERS_WIREGUARD_URL).then((res) => res.json()).then((res) => ({
    peerData: data,
    ...res
  }))
);

export const createPeerFx = createEffect<void, {ip_address: string}, Response>(
  () => fetchData(GET_IPV4_URL).then((res) => res.json())
);

export const savePeerFx = createEffect<FlatPeerType, PeerRecordType & {private_key: string}, Response>(
  (newPeer) => {
    const { public_key, ipv4, expires, label, private_key } = newPeer;

    return fetchData(
      PEERS_URL,
      {
        method: 'POST',
        body: JSON.stringify({
          type: 'wireguard',
          info_wireguard: {
            public_key
          },
          ipv4,
          expires,
          label
        })
      }
    )
      .then((res) => res.json()).then((res) => ({
        ...res,
        private_key
      }));
  }
);

export const deletePeerFx = createEffect<FlatPeerType, Response | string, Response>(
  (peer) => {
    if (!peer.created) return 'Peer deleted';

    return fetchData(
      `${PEERS_URL}/${peer.id}`,
      {
        method: 'DELETE'
      }
    );
  }
);

export const changePeerFx = createEffect<FlatPeerType, PeerType, Response>(
  (changedPeer) => {
    const { id, public_key, user_id, installation_id, session_id, ...rest } = changedPeer;

    return fetchData(
      `${PEERS_URL}/${id}`,
      {
        method: 'PUT',
        body: JSON.stringify({
          type: 'wireguard',
          info_wireguard: {
            public_key
          },
          identifiers: {
            user_id,
            installation_id,
            session_id
          },
          ...rest
        })
      }
    ).then((res) => res.json());
  }
);
