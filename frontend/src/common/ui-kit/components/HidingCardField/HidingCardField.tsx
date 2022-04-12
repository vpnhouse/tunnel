import React, { FC, useCallback, useEffect, useState } from 'react';
import { Typography } from '@material-ui/core';
import { Add, Delete, WarningRounded } from '@material-ui/icons';

import {
  TextField,
  TextButton,
  DateTimePicker,
  IconButton,
  CopyToClipboardButton,
  FileInput,
  TextArea,
  MultiTextField
} from '../index';
import { PLAIN_TEXT_FIELD } from './HidingCardField.constants';
import { PropsType } from './HidingCardField.types';
import useStyles from './HidingCardField.styles';

const HidingCardField: FC<PropsType> = ({
  isEditing,
  readonly = false,
  header = false,
  copyToClipboard = false,
  label,
  name,
  value,
  validationError,
  serverError,
  onRemoveFieldHandler,
  options = PLAIN_TEXT_FIELD,
  loadOptions
}) => {
  const classes = useStyles();

  const [isHidden, setIsHidden] = useState(!value);

  /** To expand fields with values automatically filled on back (ipv4 for example) */
  useEffect(() => {
    setIsHidden(!value);
  }, [value]);

  const showFieldHandler = useCallback(() => {
    setIsHidden(false);
  }, []);

  const deleteFieldHandler = useCallback(() => {
    onRemoveFieldHandler(name);
    setIsHidden(true);
  }, [onRemoveFieldHandler, name]);

  return (
    <>
      {isEditing && !readonly
        ? (
          <>
            {isHidden
              ? (
                <div className={classes.fieldLine}>
                  <div className={classes.leftLine} />
                  <TextButton
                    icon={Add}
                    label={label}
                    onClick={showFieldHandler}
                  />
                  <div className={classes.rightLine} />
                </div>
              )
              : (
                <div className={classes.inputLine}>
                  <>
                    {(options.type === 'TEXT' || options.type === 'TEXTAREA') && (
                      <TextField
                        {...options.textprops}
                        fullWidth
                        multiline={options.type === 'TEXTAREA'}
                        variant="outlined"
                        label={label}
                        name={name}
                        value={value}
                        helperText={validationError}
                        error={!!validationError}
                      />
                    )}

                    {(options.type === 'MULTI') && (
                      <MultiTextField
                        {...options}
                        label={label}
                        fieldName={name}
                        compoundValue={value}
                      />
                    )}

                    {options.type === 'DATETIME' && (
                      <DateTimePicker
                        {...options}
                        value={value}
                        validationError={validationError}
                      />
                    )}
                  </>
                  <div className={classes.actions}>
                    <IconButton
                      color="error"
                      onClick={deleteFieldHandler}
                      icon={Delete}
                    />
                    {copyToClipboard && <CopyToClipboardButton value={value} />}
                    {!!loadOptions && (
                      <FileInput
                        name={name}
                        {...loadOptions}
                      />
                    )}
                  </div>
                </div>
              )}
          </>
        )
        : (
          <>
            {header
              ? (
                <div className={classes.header}>
                  <Typography variant="h5">
                    {value}
                  </Typography>
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
                  <div className={classes.value}>
                    {options.type === 'TEXT' && (
                      <Typography variant="body1" className={classes.text}>
                        {value}
                      </Typography>
                    )}

                    {options.type === 'TEXTAREA' && <TextArea value={value} tableView />}

                    {options.type === 'MULTI' && (
                      <Typography variant="body1" className={classes.text}>
                        {value}
                      </Typography>
                    )}

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
              ))}
          </>
        )}
    </>
  );
};

export default HidingCardField;
