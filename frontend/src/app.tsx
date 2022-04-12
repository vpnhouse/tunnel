import React, { FC, useEffect } from 'react';
import * as ReactDOM from 'react-dom';
import './wireguard';
import { useStore } from 'effector-react';
import { Route, Switch } from 'react-router';
import { BrowserRouter } from 'react-router-dom';
import { ThemeProvider } from '@material-ui/core/styles';
import '@fontsource/roboto/400.css';
import '@fontsource/roboto/500.css';
import '@fontsource/roboto/700.css';
import '@fontsource/roboto/900.css';
import '@fontsource/roboto-mono/300.css';
import '@fontsource/roboto-mono/400.css';
import '@fontsource/roboto-mono/500.css';
import '@fontsource/roboto-mono/700.css';

import { AUTH_ROUTE, INITIAL_CONFIGURATION, MAIN_ROUTE } from '@constants/routes';
import { PrivateRoute, NotificationsBar, Dialog, GlobalLoader } from '@common/components';
import '@root/store/init';

import { Main, Auth, InitialConfiguration } from './views';
import { theme } from './app.theme';
import './app.styles.css';
import { checkConfigurationFx } from './store/initialSetup';
import { $globalLoading } from './store/globalLoading';

const App: FC = () => {
  const globalLoading = useStore($globalLoading);

  useEffect(() => {
    checkConfigurationFx();
  }, []);

  return (
    <ThemeProvider theme={theme}>
      {globalLoading && <GlobalLoader />}
      {!globalLoading && (
        <BrowserRouter>
          <Switch>
            <Route path={AUTH_ROUTE} component={Auth} />
            <Route exact path={INITIAL_CONFIGURATION} component={InitialConfiguration} />
            <PrivateRoute path={MAIN_ROUTE} component={Main} />
          </Switch>

          <NotificationsBar />
        </BrowserRouter>
      )}
      <Dialog />
    </ThemeProvider>
  );
};

const root = document.getElementById('root');

if (root) {
  root.style.height = '100%';

  ReactDOM.render(<App />, root);
}
