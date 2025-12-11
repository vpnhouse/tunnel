import { FC, useCallback, useState } from 'react';
import { DatePicker as MuiDatePicker } from '@mui/x-date-pickers/DatePicker';
import { styled } from '@mui/material/styles';

import { PropsType } from './DatePicker.types';
import { PAST_MESSAGE } from './DatePicker.constants';

const StyledDatePicker = styled(MuiDatePicker)(({ theme }) => ({
  display: 'block',
  '& .MuiInputBase-root': {
    borderRadius: 8,
    backgroundColor: '#2B3142',
    '&:hover': {
      backgroundColor: '#3B3F63'
    },
    '&.Mui-focused': {
      backgroundColor: '#3B3F63'
    }
  },
  '& .MuiInputBase-input': {
    color: theme.palette.text.primary,
    padding: '12px 14px'
  },
  '& .MuiOutlinedInput-notchedOutline': {
    borderColor: 'transparent'
  },
  '& .MuiInputLabel-root': {
    color: theme.palette.text.disabled
  },
  '& .MuiFormHelperText-root': {
    textAlign: 'end'
  },
  '& .MuiIconButton-root': {
    color: theme.palette.text.secondary
  }
})) as typeof MuiDatePicker;

const DatePicker: FC<PropsType> = ({ isEmpty = true, ...props }) => {
  const [open, setOpen] = useState(false);

  const openHandler = useCallback(() => setOpen(true), []);
  const closeHandler = useCallback(() => setOpen(false), []);

  return (
    <StyledDatePicker
      {...props}
      open={open}
      onOpen={openHandler}
      onClose={closeHandler}
      format="dd/MM/yyyy"
      slotProps={{
        textField: {
          placeholder: 'dd/mm/yyyy',
          margin: 'dense',
          helperText: props.minDate && props.value && props.value < props.minDate ? PAST_MESSAGE : undefined
        }
      }}
    />
  );
};

export default DatePicker;
