import React, { FC, useCallback, useRef, useState } from 'react';
import { useStore } from 'effector-react';
import { Typography } from '@material-ui/core';

import { PeerCard } from '@common/components';
import { Button } from '@common/ui-kit/components';
import { $peersStore, cancelCreatePeer, createPeerFx, getAllPeersFx } from '@root/store/peers';
import RefreshIcon from '@common/assets/RefreshIcon';
import AddIcon from '@common/assets/AddIcon';

import useStyles from './index.styles';

const Peers: FC = () => {
  const classes = useStyles();
  const { peers, peerToSave } = useStore($peersStore);
  const topPageRef = useRef<HTMLDivElement | null>(null);

  const [modalOpened, setModalOpened] = useState(false);

  const addPeerHandler = useCallback(() => {
    !peerToSave && createPeerFx();
    topPageRef?.current?.scrollIntoView();
    setModalOpened(true);
  }, [peerToSave]);

  const refreshPageHandler = useCallback(() => {
    getAllPeersFx();
    cancelCreatePeer();
  }, []);

  function closeModal() {
    setModalOpened(false);
  }

  return (
    <div className={classes.root}>
      <div className={classes.header}>
        <Typography variant="h1" color="textPrimary">
          Peers
        </Typography>
        <div className={classes.actions}>
          <Button
            className={classes.refreshButton}
            variant="contained"
            color="secondary"
            startIcon={<RefreshIcon />}
            onClick={refreshPageHandler}
          >
            Refresh
          </Button>
          <Button
            className={classes.addButton}
            variant="contained"
            color="primary"
            startIcon={<AddIcon />}
            onClick={addPeerHandler}
          >
            Add
          </Button>
        </div>
      </div>
      <div className={classes.main}>
        <div ref={topPageRef} />

        <PeerCard
          key="newPeer"
          peerInfo={peerToSave}
          isModal
          open={modalOpened}
          onClose={closeModal}
        />
        {peers.map((item) => <PeerCard key={item.peerInfo.id} {...item} />)}
      </div>
    </div>
  );
};

export default Peers;
