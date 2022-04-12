import React, { FC } from 'react';
import { Alert as MaterialAlert } from '@material-ui/lab';

import { PropsType } from './Alert.types';
import useStyles from './Alert.styles';

const Alert: FC<PropsType> = ({ message, ...props }) => {
  const classes = useStyles();

  return (
    <MaterialAlert
      {...props}
      classes={{
        root: classes.root,
        icon: classes.icon,
        message: classes.message,
        filledError: classes.filledError,
        filledInfo: classes.filledInfo,
        filledWarning: classes.filledWarning
      }}
      variant="filled"
    >
      {message}
    </MaterialAlert>
  );
};

export default Alert;
