import React, { FC } from 'react';
import { useLocation } from 'react-router';
import { List } from '@material-ui/core';

import { RouteMenuItem } from '@common/components';
import { PEERS_ROUTE, SETTINGS_ROUTE } from '@constants/routes';
import { getAllPeersFx } from '@root/store/peers';
import LogoutMenuItem from '@common/components/LogoutMenuItem/LogoutMenuItem';

import useStyles from './Menu.styles';
import PeersIcon from './assets/PeersIcon';
import SettingsIcon from './assets/SettingsIcon';

const Menu: FC = () => {
  const classes = useStyles();
  const { pathname } = useLocation();

  function refetchPeers() {
    getAllPeersFx();
  }

  return (
    <List
      component="nav"
      classes={{
        root: classes.root
      }}
    >
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
      <LogoutMenuItem />
    </List>
  );
};

export default Menu;
