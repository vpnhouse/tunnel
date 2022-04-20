import React, { FC, useState } from 'react';
import { Typography, Tooltip, Switch, FormControlLabel } from '@material-ui/core';
import { HelpOutlineRounded, WarningRounded } from '@material-ui/icons';

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
  faq = false,
  faqText,
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
              <div className={faq ? classes.field__faq : ''}>
                {faq
                  ? (
                    <>
                      <div className={classes.field__faq_wrap}>
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
                        <Tooltip placement="right-start" title={faqText || ''}>
                          <HelpOutlineRounded className={classes.field__faq_icon} />
                        </Tooltip>
                      </div>
                      {disableControl
                        ? (
                          <div className={classes.disable__control}>
                            <FormControlLabel
                              control={
                                <Switch onChange={handleGameClick} />
                              }
                              label="Change field"
                            />
                          </div>
                        )
                        : ''
                      }
                    </>
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
                }
              </div>
            )}

            {options.type === 'DATETIME' && (
              <DateTimePicker
                {...options}
                value={value}
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
