import { FC } from 'react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import CircularProgress from '@mui/material/CircularProgress';
import { styled } from '@mui/material/styles';

const LoaderSection = styled(Box)(({ theme }) => ({
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  justifyContent: 'center',
  height: '100vh',
  width: '100%',
  backgroundColor: theme.palette.background.default
}));

const GlobalLoader: FC = () => {
  return (
    <LoaderSection>
      <CircularProgress size={48} color="primary" />
      <Typography
        variant="h6"
        sx={{ mt: 2, color: 'text.primary' }}
      >
        Loading...
      </Typography>
    </LoaderSection>
  );
};

export default GlobalLoader;
