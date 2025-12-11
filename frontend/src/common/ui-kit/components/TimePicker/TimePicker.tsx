import { FC } from 'react';
import { TimePicker as MuiTimePicker } from '@mui/x-date-pickers/TimePicker';
import { styled } from '@mui/material/styles';

import { PropsType } from './TimePicker.types';

const StyledTimePicker = styled(MuiTimePicker)(({ theme }) => ({
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
})) as typeof MuiTimePicker;

const TimePicker: FC<PropsType> = ({ isEmpty, ...props }) => {
  return (
    <StyledTimePicker
      {...props}
      ampm={false}
      format="HH:mm"
      views={['hours', 'minutes']}
      slotProps={{
        textField: {
          placeholder: 'hh:mm',
          margin: 'dense'
        }
      }}
    />
  );
};

export default TimePicker;
