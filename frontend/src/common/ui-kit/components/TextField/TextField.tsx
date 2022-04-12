import React, { FC } from 'react';
import { TextField as MaterialTextField } from '@material-ui/core';

import useStyles from './TextField.styles';
import { PropsType } from './TextField.types';

const TextField: FC<PropsType> = ({ endAdornment, ...props }) => {
  const classes = useStyles(props);

  return (
    <MaterialTextField
      className={classes.root}
      {...props}
      variant="filled"
      margin="dense"
      InputLabelProps={{
        classes: {
          root: classes.labelRoot,
          filled: classes.inputLabelFilled
        }
      }}
      InputProps={{
        classes: {
          root: classes.inputRoot,
          adornedEnd: classes.adornedEnd,
          input: classes.input,
          marginDense: classes.inputMarginDense
        },
        endAdornment,
        disableUnderline: true
      }}
      FormHelperTextProps={{
        classes: {
          root: classes.helperText
        }
      }}
    />
  );
};

export default TextField;
