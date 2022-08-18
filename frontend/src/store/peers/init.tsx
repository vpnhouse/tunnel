import React from 'react';

import PersonalPrivateData from '@root/common/components/PersonalPrivateData/PersonalPrivateData';
import PersonalPrivateDataAction from '@root/common/components/PersonalPrivateData/PersonalPrivateDataAction';

import {
  $peersStore,
  setPeers,
  createPeer,
  cancelCreatePeer,
  savePeer,
  deletePeer,
  changePeer,
  getAllPeersFx,
  savePeerFx,
  deletePeerFx,
  changePeerFx,
  setIsEditing,
  getPeersWireguardFx,
  createPeerFx
} from './index';
import { addNotification, showServerErrorFx } from '../notifications';
import { openDialog } from '../dialogs';

$peersStore
  .on(setPeers, (store, peersList) => ({
    ...store,
    peers: peersList
      .sort((a, b) => {
        if (a.peer.updated && b.peer.updated) return Date.parse(b.peer.updated) - Date.parse(a.peer.updated);

        return 0;
      })
      .map((peerRecord) => {
        const { id, peer } = peerRecord;
        const { info_wireguard, identifiers, ...rest } = peer;
        return {
          peerInfo: {
            id,
            ...info_wireguard,
            ...identifiers,
            ...rest
          },
          isEditing: false
        };
      })
  }))
  .on(createPeer, (store, ipv4) => ({
    ...store,
    peerToSave: {
      id: 0,
      ipv4
    }
  }))
  .on(cancelCreatePeer, (store) => ({
    ...store,
    peerToSave: null
  }))
  .on(savePeer, (store, newPeer) => ({
    ...store,
    peers: [newPeer, ...store.peers],
    peerToSave: null
  }))
  .on(deletePeer, (store, id) => ({
    ...store,
    peers: store.peers.filter((item) => item.peerInfo.id !== id)
  }))
  .on(changePeer, (store, changedPeer) => ({
    ...store,
    peers: store.peers.map((item) => (
      item.peerInfo.id === changedPeer.peerInfo.id ? changedPeer : item
    ))
  }))
  .on(setIsEditing, (store, { id, isEditing }) => ({
    ...store,
    peers: store.peers.map((item) => (
      item.peerInfo.id === id
        ? {
          ...item,
          isEditing
        }
        : item
    ))
  }));

getAllPeersFx.doneData.watch((result) => {
  setPeers(result);
});

getAllPeersFx.failData.watch((error) => {
  showServerErrorFx(error);
});

createPeerFx.doneData.watch((res) => {
  createPeer(res.ip_address);
});

createPeerFx.failData.watch(
  (error) => error.json().then((err) => {
    addNotification({
      type: 'error',
      prefix: 'serverError',
      message: err.error
    });
  })
);

savePeerFx.watch((params) => {
  /** If peer wasn't saved, delete it from store, it will be saved again later with proper id */
  !params.created && deletePeer(params.id);
});


getPeersWireguardFx.doneData.watch((res) => {
  const template = `
  [Interface]
  PrivateKey = ${res.peerData.private_key}
  Address = ${res.peerData.ipv4}
  DNS = ${res.dns.join(', ')}

  [Peer]
  PublicKey = ${res.server_public_key}
  AllowedIPs =  ${res.allowed_ips.join(', ')}
  Endpoint = ${res.server_ipv4}:${res.server_port}
  PersistentKeepalive = ${res.keepalive}
`;

  const textBlob = new Blob([template.slice(1)], { type: 'text/plain' });
  const fileLink = window.URL.createObjectURL(textBlob);

  openDialog({
    title: `${res.peerData.label} configuration`,
    message: <PersonalPrivateData value={template} />,
    actionComponent: <PersonalPrivateDataAction fileLink={fileLink} />
  });
});

getPeersWireguardFx.failData.watch(
  (error) => error.json().then((err) => {
    addNotification({
      type: 'error',
      prefix: 'serverError',
      message: err.error
    });
  })
);

savePeerFx.doneData.watch((result) => {
  const { id, peer, private_key } = result;
  const { info_wireguard, identifiers, ...rest } = peer;

  savePeer({
    peerInfo: {
      id,
      info_wireguard: {
        ...info_wireguard,
        private_key
      },
      ...identifiers,
      ...rest
    },
    isEditing: false
  });
  getPeersWireguardFx({
    private_key,
    label: rest.label,
    ipv4: peer.ipv4 as string
  });
});

savePeerFx.fail.watch(({ params, error }) => {
  error.json().then((errorDetails) => {
    const field = errorDetails.field || 'common';
    const serverError = { [field]: `${errorDetails.error} ${errorDetails.details || ''}` };
    const { id, ...rest } = params;

    /** Save in store with server error info and negative id (to be sure there is no such id in DB) */
    savePeer({
      peerInfo: {
        id: id || -(Date.now()),
        ...rest
      },
      serverError,
      isEditing: false
    });
  });
});

deletePeerFx.done.watch(({ params }) => {
  deletePeer(params.id);
  addNotification({
    type: 'info',
    prefix: 'peerDeleteInfo',
    message: `Peer ${params.label || ''} was removed`
  });
});

deletePeerFx.failData.watch((error) => {
  showServerErrorFx(error);
});

changePeerFx.done.watch(({ params, result }) => {
  const { id } = params;
  const { info_wireguard, identifiers, ...rest } = result;

  changePeer({
    peerInfo: {
      id,
      ...info_wireguard,
      ...identifiers,
      ...rest
    },
    isEditing: false
  });
});

changePeerFx.fail.watch(({ params, error }) => {
  error.json().then((errorDetails) => {
    const field = errorDetails.field || 'common';
    const serverError = { [field]: `${errorDetails.error} ${errorDetails.details || ''}` };
    changePeer({
      peerInfo: params,
      serverError,
      isEditing: false
    });
  });
});
