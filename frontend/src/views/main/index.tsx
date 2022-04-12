import React, { FC } from 'react';
import { Route, Switch, Redirect } from 'react-router';

import { NavigationPanel } from '@common/components';
import { Peers, Settings, TrustedKeys } from '@root/views';
import { PEERS_ROUTE, SETTINGS_ROUTE, TRUSTED_ROUTE } from '@constants/routes';

import useStyles from './index.styles';

const Main: FC = () => {
  const classes = useStyles();

  return (
    <div className={classes.root}>
      <NavigationPanel />
      <div className={classes.content}>
        <Switch>
          <Route exact path={PEERS_ROUTE} component={Peers} />
          <Route exact path={SETTINGS_ROUTE} component={Settings} />
          <Route exact path={TRUSTED_ROUTE} component={TrustedKeys} />
          <Route exact path="/">
            <Redirect to={PEERS_ROUTE} />
          </Route>
        </Switch>
      </div>
    </div>
  );
};

export default Main;
