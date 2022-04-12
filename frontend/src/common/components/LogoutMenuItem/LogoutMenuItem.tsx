import React, { FC, useCallback } from 'react';
import { ListItem, ListItemIcon, ListItemText } from '@material-ui/core';

import { openDialog } from '@root/store/dialogs';
import { logout } from '@root/store/auth';
import LogoutIcon from '@common/components/LogoutMenuItem/assets/LogoutIcon';

import useStyles from './LogoutMenuItem.styles';

const LogoutMenuItem: FC = () => {
  const classes = useStyles();

  const logoutBtnClickHandler = useCallback(() => {
    openDialog({
      title: 'Confirm log out',
      message: 'Do you want to log out from service?',
      successButtonTitle: 'Log out',
      successButtonHandler: logout
    });
  }, []);

  return (
    <ListItem
      button
      onClick={logoutBtnClickHandler}
      classes={{
        root: classes.itemRoot,
        selected: classes.itemSelected
      }}
    >
      <ListItemIcon
        classes={{
          root: classes.listItemIconRoot
        }}
      >
        <LogoutIcon className={classes.iconRoot} />
      </ListItemIcon>
      <ListItemText
        classes={{
          primary: classes.primaryText
        }}
        primary="Log out"
      />
    </ListItem>
  );
};

export default LogoutMenuItem;
