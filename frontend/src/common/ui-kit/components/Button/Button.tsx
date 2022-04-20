import React, { FC } from 'react';
import { Button as MaterialButton, ButtonProps, CircularProgress } from '@material-ui/core';

import useStyles from './Button.styles';

interface Props extends ButtonProps {
  isLoading?: boolean;
}

const Button: FC<Props> = ({ children, isLoading, disabled, ...rest }) => {
  const classes = useStyles();

  return (
    <MaterialButton
      {...rest}
      disabled={disabled || isLoading}
      classes={{
        root: classes.root,
        containedPrimary: classes.containedPrimary,
        containedSecondary: classes.containedSecondary,
        text: classes.text,
        label: classes.label
      }}
    >
      {children}

      {isLoading && <CircularProgress className={classes.spinner} />}
    </MaterialButton>
  );
};

export default Button;
