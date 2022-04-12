import React, { FC } from 'react';
import { KeyboardTimePicker } from '@material-ui/pickers';

import useStyles from './TimePicker.styles';
import { PropsType } from './TimePicker.types';

const TimePicker: FC<PropsType> = ({ isEmpty, ...props }) => {
  const classes = useStyles({ isEmpty });

  return (
    <KeyboardTimePicker
      {...props}
      variant="inline"
      inputVariant="outlined"
      ampm={false}
      format="HH:mm"
      views={['hours', 'minutes']}
      className={classes.root}
      margin="dense"
      placeholder="hh:mm"
      InputLabelProps={{
        classes: {
          root: classes.labelRoot,
          outlined: classes.inputLabelFilled
        }
      }}
      InputProps={{
        classes: {
          root: classes.inputRoot,
          notchedOutline: classes.notchedOutline,
          input: classes.input,
          marginDense: classes.inputMarginDense,
          adornedEnd: classes.adornedEnd
        }
      }}
      FormHelperTextProps={{
        classes: {
          root: classes.helperText
        }
      }}
    />
  );
};

export default TimePicker;
