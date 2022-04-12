import React, { FC, useCallback, useState } from 'react';
import { MuiPickersUtilsProvider } from '@material-ui/pickers';
import DateFnsUtils from '@date-io/date-fns';
import { Typography } from '@material-ui/core';

import { DatePicker, TimePicker } from '@common/ui-kit/components';

import { PropsType } from './DateTimePicker.types';
import useStyles from './DateTimePicker.styles';


const DateTimePicker: FC<PropsType> = ({
  dateLabel,
  dateName,
  timeLabel,
  timeName,
  value,
  validationError,
  onChangeHandler,
  datePickerProps,
  timePickerProps
}) => {
  const classes = useStyles();
  const [date, setDate] = useState<string | null>(value);
  const [time, setTime] = useState<string | null>(value);

  const dateChangeHandler = useCallback((input) => {
    setDate(input);

    const newDate = input || null;
    const newTime = time ? new Date(time) : null;

    onChangeHandler(newDate, newTime);
  }, [time, onChangeHandler]);

  const timeChangeHandler = useCallback((input) => {
    setTime(input);

    const newDate = date ? new Date(date) : null;
    const newTime = input || null;

    onChangeHandler(newDate, newTime);
  }, [date, onChangeHandler]);

  return (
    <div className={classes.root}>
      <div className={classes.pickers}>
        <MuiPickersUtilsProvider utils={DateFnsUtils}>
          <DatePicker
            {...datePickerProps}
            isEmpty={!value}
            label={dateLabel}
            name={dateName}
            value={date || null}
            onChange={dateChangeHandler}
          />
          <TimePicker
            {...timePickerProps}
            isEmpty={!value}
            label={timeLabel}
            name={timeName}
            value={time || null}
            onChange={timeChangeHandler}
          />
        </MuiPickersUtilsProvider>
      </div>
      {!!validationError && (
        <Typography className={classes.validationError} variant="caption" component="p">
          {validationError}
        </Typography>
      )}
    </div>

  );
};

export default DateTimePicker;
