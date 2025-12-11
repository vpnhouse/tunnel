import { FC, useState } from 'react';
import Typography from '@mui/material/Typography';
import Switch from '@mui/material/Switch';
import FormControlLabel from '@mui/material/FormControlLabel';
import Box from '@mui/material/Box';
import WarningRounded from '@mui/icons-material/WarningRounded';
import { styled } from '@mui/material/styles';

import { CopyToClipboardButton, TextField, FileInput, TextArea, DateTimePicker } from '../index';
import { PropsType, TextFieldType } from './CardField.types';
import { PLAIN_TEXT_FIELD } from './CardField.constants';

const AreaRoot = styled(Box)({
  display: 'flex',
  alignItems: 'flex-start'
});

const FieldWithControl = styled(Box)({
  display: 'flex',
  flexDirection: 'column',
  width: '100%'
});

const DisableControl = styled(Box)({
  marginTop: 8
});

const Actions = styled(Box)({
  display: 'flex',
  marginLeft: 8
});

const TextBlock = styled(Box)({
  marginBottom: 16
});

const Caption = styled(Box)(({ theme }) => ({
  display: 'flex',
  alignItems: 'center',
  gap: 4,
  marginBottom: 4,
  fontSize: theme.typography.caption.fontSize,
  color: theme.palette.text.secondary
}));

const Value = styled(Box)({
  wordBreak: 'break-all'
});

const ErrorText = styled(Typography)(({ theme }) => ({
  color: theme.palette.error.main,
  marginTop: 4
}));

const DateTimePickerWrapper = styled(Box)({
  display: 'flex',
  gap: 12
});

interface StyleProps {
  tableView?: boolean;
}

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
  const { type } = options;
  const [disabled, setDisabled] = useState(true);

  function handleGameClick() {
    setDisabled(!disabled);
  }

  return (
    <>
      {isEditing && !readonly
        ? (
          <Box component={type === 'TEXTAREA' ? AreaRoot : 'div'}>
            {(type === 'TEXT' || type === 'TEXTAREA') && (
              disableControl
                ? (
                  <FieldWithControl>
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
                    <DisableControl>
                      <FormControlLabel
                        control={
                          <Switch onChange={handleGameClick} />
                        }
                        label="Change field"
                      />
                    </DisableControl>
                  </FieldWithControl>
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
                classNames={{ pickers: undefined }}
                validationError={validationError ?? ''}
              />
            )}

            <Actions>
              {copyToClipboard && <CopyToClipboardButton value={value} />}
              {!!loadOptions && (
                <FileInput
                  name={name}
                  {...loadOptions}
                />
              )}
            </Actions>
          </Box>
        )
        : (!!value && (
          <TextBlock>
            <Caption
              sx={{ color: serverError ? 'error.main' : 'text.secondary' }}
            >
              {!!serverError && <WarningRounded fontSize="small" />}
              {label}
            </Caption>

            <Box sx={tableView ? { wordBreak: 'break-all' } : undefined}>
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
                <ErrorText>
                  {serverError}
                </ErrorText>
              )}
            </Box>
          </TextBlock>
        ))
      }
    </>
  );
};

export default CardField;
