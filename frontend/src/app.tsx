import { FC, useEffect } from 'react';
import { createRoot } from 'react-dom/client';
import './wireguard';
import { useUnit } from 'effector-react';
import { Route, Routes, BrowserRouter } from 'react-router-dom';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
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
  const globalLoading = useUnit($globalLoading);

  useEffect(() => {
    checkConfigurationFx();
  }, []);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      {globalLoading && <GlobalLoader />}
      {!globalLoading && (
        <BrowserRouter>
          <Routes>
            <Route path={`${AUTH_ROUTE}/*`} element={<Auth />} />
            <Route path={INITIAL_CONFIGURATION} element={<InitialConfiguration />} />
            <Route path={`${MAIN_ROUTE}/*`} element={<PrivateRoute><Main /></PrivateRoute>} />
          </Routes>

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
  createRoot(root).render(<App />);
}
