import React, { FC, useCallback, useState } from 'react';
import { KeyboardDatePicker } from '@material-ui/pickers';

import useStyles from './DatePicker.styles';
import { PropsType } from './DatePicker.types';
import { PAST_MESSAGE } from './DatePicker.constants';

const DatePicker: FC<PropsType> = ({ isEmpty = true, ...props }) => {
  const classes = useStyles({ isEmpty });
  const [open, setOpen] = useState(false);

  const openHandler = useCallback(() => setOpen(true), []);
  const closeHandler = useCallback(() => setOpen(false), []);

  return (
    <KeyboardDatePicker
      {...props}
      open={open}
      disableToolbar
      variant="inline"
      inputVariant="outlined"
      placeholder="dd/mm/yyyy"
      format="dd/MM/yyyy"
      onOpen={openHandler}
      onAccept={closeHandler}
      onClose={closeHandler}
      className={classes.root}
      margin="dense"
      minDateMessage={PAST_MESSAGE}
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
      PopoverProps={{
        PaperProps: {
          classes: {
            root: classes.paper
          }
        }
      }}
    />
  );
};

export default DatePicker;
