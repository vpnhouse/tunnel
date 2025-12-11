import { FC, useCallback, useRef, useState } from 'react';
import { useUnit } from 'effector-react';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import { useTheme } from '@mui/material/styles';

import { PeerCard } from '@common/components';
import { Button } from '@common/ui-kit/components';
import { $peersStore, cancelCreatePeer, createPeerFx, getAllPeersFx } from '@root/store/peers';
import RefreshIcon from '@common/assets/RefreshIcon';
import AddIcon from '@common/assets/AddIcon';

const Peers: FC = () => {
  const theme = useTheme();
  const { peers, peerToSave } = useUnit($peersStore);
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
    <Box
      sx={{
        height: '100%',
        display: 'flex',
        paddingRight: '64px',
        flexDirection: 'column',
        '@media(max-width: 1359px)': {
          paddingRight: '32px'
        }
      }}
    >
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          width: '100%',
          margin: '36px 0',
          userSelect: 'none'
        }}
      >
        <Typography variant="h1" color="textPrimary">
          Peers
        </Typography>
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'flex-end',
            '& button:not(:last-child)': {
              marginRight: '12px'
            },
            '& button': {
              padding: '0 28px'
            }
          }}
        >
          <Button
            sx={{
              width: 160,
              '& svg': { height: 16, width: 16 }
            }}
            variant="contained"
            color="secondary"
            startIcon={<RefreshIcon />}
            onClick={refreshPageHandler}
          >
            Refresh
          </Button>
          <Button
            sx={{
              width: 160,
              '& svg': { height: 14, width: 14 }
            }}
            variant="contained"
            color="primary"
            startIcon={<AddIcon />}
            onClick={addPeerHandler}
          >
            Add
          </Button>
        </Box>
      </Box>
      <Box sx={{ height: '100%', overflow: 'auto', width: '100%' }}>
        <div ref={topPageRef} />
        <Box sx={{ display: 'flex' }}>
          <Box
            sx={{
              display: 'flex',
              alignItems: 'flex-start',
              flexWrap: 'wrap',
              width: '100%',
              maxWidth: '1080px'
            }}
          >
            {peerToSave && (
              <PeerCard
                key="newPeer"
                peerInfo={peerToSave}
                isModal
                open={modalOpened}
                onClose={closeModal}
              />
            )}
            {peers.map((item) => <PeerCard key={item.peerInfo.id} {...item} />)}
          </Box>
        </Box>
      </Box>
    </Box>
  );
};

export default Peers;
