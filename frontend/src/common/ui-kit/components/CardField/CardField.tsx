import React, { FC, useState } from 'react';
import { Typography, Switch, FormControlLabel } from '@material-ui/core';
import { WarningRounded } from '@material-ui/icons';

import { CopyToClipboardButton, TextField, FileInput, TextArea, DateTimePicker } from '../index';
import { PropsType, TextFieldType } from './CardField.types';
import useStyles from './CardField.styles';
import { PLAIN_TEXT_FIELD } from './CardField.constants';

const CardField: FC<PropsType> = ({
  isEditing,
  readonly = false,
  copyToClipboard = false,
  tableView = false,
  label,
  name,
  value,
  validationError,
  serverError,
  options = PLAIN_TEXT_FIELD,
  isDisable = false,
  disableControl = false,
  loadOptions
}) => {
  const classes = useStyles({ tableView });
  const { type } = options;
  const [disabled, setDisabled] = useState(true);

  function handleGameClick() {
    setDisabled(!disabled);
  }

  return (
    <>
      {isEditing && !readonly
        ? (
          <div className={type === 'TEXTAREA' ? classes.areaRoot : ''}>
            {(type === 'TEXT' || type === 'TEXTAREA') && (
              disableControl
                ? (
                  <div className={classes.field__withControl}>
                    <TextField
                      {...(options as TextFieldType).textprops}
                      fullWidth
                      multiline={options.type === 'TEXTAREA'}
                      variant="outlined"
                      label={label}
                      name={name}
                      value={value}
                      helperText={validationError}
                      error={!!validationError}
                      disabled={disabled}
                    />
                    <div className={classes.disable__control}>
                      <FormControlLabel
                        control={
                          <Switch onChange={handleGameClick} />
                        }
                        label="Change field"
                      />
                    </div>
                  </div>
                )
                : (
                  <TextField
                    {...(options as TextFieldType).textprops}
                    fullWidth
                    multiline={options.type === 'TEXTAREA'}
                    variant="outlined"
                    label={label}
                    name={name}
                    value={value}
                    helperText={validationError}
                    error={!!validationError}
                    disabled={isDisable}
                  />
                )
            )}

            {options.type === 'DATETIME' && (
              <DateTimePicker
                {...options}
                value={value}
                classNames={{ pickers: classes.dateTimePicker }}
                validationError={validationError}
              />
            )}

            <div className={classes.actions}>
              {copyToClipboard && <CopyToClipboardButton value={value} />}
              {!!loadOptions && (
                <FileInput
                  name={name}
                  {...loadOptions}
                />
              )}
            </div>
          </div>
        )
        : (!!value && (
          <div className={classes.textBlock}>
            <Typography
              className={classes.caption}
              variant="caption"
              component="div"
              color={serverError ? 'error' : 'textSecondary'}
            >
              {!!serverError && <WarningRounded fontSize="small" />}
              {label}
            </Typography>

            <div className={tableView ? classes.value : ''}>
              {type === 'TEXT' && (
                <Typography variant="body1">
                  {value}
                </Typography>
              )}

              {type === 'TEXTAREA' && <TextArea value={value} tableView={tableView} />}

              {options.type === 'DATETIME' && (
              <Typography variant="body1">
                {new Date(value).toString().split(' (')[0]}
              </Typography>
              )}

              {!!serverError && (
                <Typography className={classes.error} variant="caption" component="p">
                  {serverError}
                </Typography>
              )}
            </div>
          </div>
        ))
      }
    </>
  );
};

export default CardField;
