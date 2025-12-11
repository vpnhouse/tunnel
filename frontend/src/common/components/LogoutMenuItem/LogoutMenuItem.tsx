import { FC, useCallback } from 'react';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import { styled } from '@mui/material/styles';

import { openDialog } from '@root/store/dialogs';
import { logout } from '@root/store/auth';
import LogoutIcon from '@common/components/LogoutMenuItem/assets/LogoutIcon';

const StyledListItemButton = styled(ListItemButton)(({ theme }) => ({
  padding: '12px 24px',
  marginTop: 'auto',
  borderTop: `1px solid ${theme.palette.divider}`,
  '&:hover': {
    backgroundColor: 'rgba(255, 255, 255, 0.08)'
  },
  [theme.breakpoints.down('lg')]: {
    padding: '12px',
    justifyContent: 'center'
  }
}));

const StyledListItemIcon = styled(ListItemIcon)(({ theme }) => ({
  minWidth: 40,
  color: theme.palette.error.main,
  [theme.breakpoints.down('lg')]: {
    minWidth: 'auto'
  }
}));

const StyledLogoutIcon = styled(LogoutIcon)({
  fontSize: 24
});

const StyledListItemText = styled(ListItemText)(({ theme }) => ({
  '& .MuiListItemText-primary': {
    fontSize: 14,
    fontWeight: 500,
    color: theme.palette.error.main
  },
  [theme.breakpoints.down('lg')]: {
    display: 'none'
  }
}));

const LogoutMenuItem: FC = () => {
  const logoutBtnClickHandler = useCallback(() => {
    openDialog({
      title: 'Confirm log out',
      message: 'Do you want to log out from service?',
      successButtonTitle: 'Log out',
      successButtonHandler: logout
    });
  }, []);

  return (
    <StyledListItemButton onClick={logoutBtnClickHandler}>
      <StyledListItemIcon>
        <StyledLogoutIcon />
      </StyledListItemIcon>
      <StyledListItemText primary="Log out" />
    </StyledListItemButton>
  );
};

export default LogoutMenuItem;
