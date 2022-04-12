import React, { FC } from 'react';
import { Icon } from '@material-ui/core';

import { PropsType } from './TextButton.types';
import useStyles from './TextButton.styles';

const TextButton: FC<PropsType> = ({ icon, label, onClick }) => {
  const classes = useStyles();

  return (
    <div
      className={classes.root}
      onClick={onClick}
    >
      <Icon fontSize="small" component={icon} />
      <span className={classes.label}>
        {label}
      </span>
    </div>
  );
};

export default TextButton;
