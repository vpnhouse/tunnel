import React, { FC } from 'react';
import { useStore } from 'effector-react';
import { Route, Redirect } from 'react-router-dom';

import { $authStore } from '@root/store/auth';
import { AUTH_ROUTE, INITIAL_CONFIGURATION } from '@constants/routes';
import { $initialSetup } from '@root/store/initialSetup';

import { PropsType } from './PrivateRoute.types';

const PrivateRoute: FC<PropsType> = (props) => {
  const isAuthenticated = useStore($authStore);
  const isInitialConfigurateDone = useStore($initialSetup);

  const { component, ...rest } = props;

  if (!isInitialConfigurateDone) {
    return <Redirect to={INITIAL_CONFIGURATION} />;
  }
  return isAuthenticated
    ? <Route component={component} {...rest} />
    : <Redirect to={AUTH_ROUTE} />;
};

export default PrivateRoute;
