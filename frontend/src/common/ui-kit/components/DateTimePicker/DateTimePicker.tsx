import { FC, useCallback, useState } from 'react';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import clsx from 'clsx';

import { DatePicker, TimePicker } from '@common/ui-kit/components';

import { PropsType } from './DateTimePicker.types';

const DateTimePicker: FC<PropsType> = ({
  dateLabel,
  dateName,
  timeLabel,
  timeName,
  value,
  validationError,
  onChangeHandler,
  datePickerProps,
  timePickerProps,
  classNames
}) => {
  const [date, setDate] = useState<Date | null>(value ? new Date(value) : null);
  const [time, setTime] = useState<Date | null>(value ? new Date(value) : null);

  const dateChangeHandler = useCallback((input: Date | null) => {
    setDate(input);
    const newTime = time ? new Date(time) : null;
    onChangeHandler(input, newTime);
  }, [time, onChangeHandler]);

  const timeChangeHandler = useCallback((input: Date | null) => {
    setTime(input);
    const newDate = date ? new Date(date) : null;
    onChangeHandler(newDate, input);
  }, [date, onChangeHandler]);

  return (
    <Box className={clsx(classNames?.root)} sx={{ marginBottom: 2 }}>
      <Box
        className={clsx(classNames?.pickers)}
        sx={{ display: 'flex', gap: 2 }}
      >
        <LocalizationProvider dateAdapter={AdapterDateFns}>
          <DatePicker
            {...datePickerProps}
            isEmpty={!value}
            label={dateLabel}
            value={date}
            onChange={dateChangeHandler}
          />
          <TimePicker
            {...timePickerProps}
            isEmpty={!value}
            label={timeLabel}
            value={time}
            onChange={timeChangeHandler}
          />
        </LocalizationProvider>
      </Box>
      {!!validationError && (
        <Typography
          variant="caption"
          component="p"
          sx={{ color: 'error.main', textAlign: 'end', mt: 0.5 }}
        >
          {validationError}
        </Typography>
      )}
    </Box>
  );
};

export default DateTimePicker;
