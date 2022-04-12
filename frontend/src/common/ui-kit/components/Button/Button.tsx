import React, { FC } from 'react';
import { Button as MaterialButton, ButtonProps } from '@material-ui/core';

import useStyles from './Button.styles';

const Button: FC<ButtonProps> = (props) => {
  const classes = useStyles();

  return (
    <MaterialButton
      {...props}
      classes={{
        root: classes.root,
        containedPrimary: classes.containedPrimary,
        containedSecondary: classes.containedSecondary,
        text: classes.text,
        label: classes.label
      }}
    />
  );
};

export default Button;
