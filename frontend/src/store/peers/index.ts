import { createStore, createEffect, createEvent } from 'effector';

import { GET_IPV4, PEERS, GET_WIREGUARD } from '@constants/apiPaths';

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
  () => fetchData(PEERS).then((res) => res.json())
);

export const getPeersWireguardFx = createEffect<{private_key: string; label?: string | null; ipv4: string}, PeersWireguard, Response>(
  (data) => fetchData(GET_WIREGUARD).then((res) => res.json()).then((res) => ({
    peerData: data,
    ...res
  }))
);

export const createPeerFx = createEffect<void, {ip_address: string}, Response>(
  () => fetchData(GET_IPV4).then((res) => res.json())
);

export const savePeerFx = createEffect<FlatPeerType, PeerRecordType & {private_key: string}, Response>(
  (newPeer) => {
    const { public_key, ipv4, expires, label, private_key } = newPeer;

    return fetchData(
      PEERS,
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
      `${PEERS}/${peer.id}`,
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
      `${PEERS}/${id}`,
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
