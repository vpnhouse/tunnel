import { FC } from 'react';
import { useLocation } from 'react-router';
import List from '@mui/material/List';
import { styled } from '@mui/material/styles';

import { RouteMenuItem } from '@common/components';
import { PEERS_ROUTE, SETTINGS_ROUTE } from '@constants/routes';
import { getAllPeersFx } from '@root/store/peers';
import LogoutMenuItem from '@common/components/LogoutMenuItem/LogoutMenuItem';
import GlobalStatsBar from '@common/components/Menu/GlobalStatsBar';

import PeersIcon from './assets/PeersIcon';
import SettingsIcon from './assets/SettingsIcon';

const StyledList = styled(List)(({ theme }) => ({
  display: 'flex',
  flexDirection: 'column',
  flexGrow: 1,
  padding: '24px 0'
}));

const Menu: FC = () => {
  const { pathname } = useLocation();

  function refetchPeers() {
    getAllPeersFx();
  }

  return (
    <StyledList>
      <RouteMenuItem
        selected={pathname === PEERS_ROUTE}
        icon={PeersIcon}
        route={PEERS_ROUTE}
        pageTitle="Peers"
        onClick={refetchPeers}
      />
      <RouteMenuItem
        selected={pathname === SETTINGS_ROUTE}
        icon={SettingsIcon}
        route={SETTINGS_ROUTE}
        pageTitle="Settings"
      />
      <GlobalStatsBar />
      <LogoutMenuItem />
    </StyledList>
  );
};

export default Menu;
