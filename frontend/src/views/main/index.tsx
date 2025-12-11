import { FC } from 'react';
import { Route, Routes, Navigate } from 'react-router-dom';
import Box from '@mui/material/Box';
import { useTheme } from '@mui/material/styles';

import { NavigationPanel } from '@common/components';
import { Peers, Settings, TrustedKeys } from '@root/views';
import { PEERS_ROUTE, SETTINGS_ROUTE, TRUSTED_ROUTE } from '@constants/routes';

const Main: FC = () => {
  const theme = useTheme();

  return (
    <Box
      sx={{
        height: '100%',
        display: 'flex',
        backgroundColor: theme.palette.background.default,
        overflowX: 'auto'
      }}
    >
      <NavigationPanel />
      <Box sx={{ flex: '1 1 0' }}>
        <Routes>
          <Route path={PEERS_ROUTE.replace('/', '')} element={<Peers />} />
          <Route path={SETTINGS_ROUTE.replace('/', '')} element={<Settings />} />
          <Route path={TRUSTED_ROUTE.replace('/', '')} element={<TrustedKeys />} />
          <Route path="/" element={<Navigate to={PEERS_ROUTE} replace />} />
          <Route path="*" element={<Navigate to={PEERS_ROUTE} replace />} />
        </Routes>
      </Box>
    </Box>
  );
};

export default Main;
