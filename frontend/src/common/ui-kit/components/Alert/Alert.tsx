import { FC } from 'react';
import Alert from '@mui/material/Alert';
import { styled } from '@mui/material/styles';
import { PropsType } from './Alert.types';

const StyledAlert = styled(Alert)(({ theme }) => ({
  marginBottom: '10px',
  padding: '11px 32px',
  color: theme.palette.common.white,
  ...theme.typography.subtitle1,
  width: '760px',
  '& .MuiAlert-icon': {
    display: 'none'
  },
  '& .MuiAlert-message': {
    padding: 0,
    display: 'flex',
    alignItems: 'center'
  },
  '&.MuiAlert-filledError': {
    backgroundColor: theme.palette.error.main
  },
  '&.MuiAlert-filledInfo': {
    backgroundColor: theme.palette.info.main
  },
  '&.MuiAlert-filledWarning': {
    backgroundColor: theme.palette.info.main
  }
}));

const CustomAlert: FC<PropsType> = ({ message, ...props }) => {
  return (
    <StyledAlert {...props} variant="filled">
      {message}
    </StyledAlert>
  );
};

export default CustomAlert;
