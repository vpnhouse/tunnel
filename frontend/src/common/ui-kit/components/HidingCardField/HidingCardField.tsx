import { FC, useCallback, useEffect, useState } from 'react';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Add from '@mui/icons-material/Add';
import Delete from '@mui/icons-material/Delete';
import WarningRounded from '@mui/icons-material/WarningRounded';
import { styled } from '@mui/material/styles';

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

const FieldLine = styled(Box)({
  display: 'flex',
  alignItems: 'center',
  margin: '16px 0'
});

const Line = styled(Box)(({ theme }) => ({
  flex: 1,
  height: 1,
  backgroundColor: theme.palette.divider
}));

const InputLine = styled(Box)({
  display: 'flex',
  alignItems: 'flex-start',
  marginBottom: 16
});

const Actions = styled(Box)({
  display: 'flex',
  marginLeft: 8,
  gap: 4
});

const Header = styled(Box)({
  marginBottom: 16
});

const TextBlock = styled(Box)({
  marginBottom: 16
});

const Caption = styled(Typography)({
  display: 'flex',
  alignItems: 'center',
  gap: 4,
  marginBottom: 4
});

const Value = styled(Box)({
  wordBreak: 'break-all'
});

const ErrorText = styled(Typography)(({ theme }) => ({
  color: theme.palette.error.main,
  marginTop: 4
}));

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
                <FieldLine>
                  <Line />
                  <TextButton
                    icon={Add}
                    label={label}
                    onClick={showFieldHandler}
                  />
                  <Line />
                </FieldLine>
              )
              : (
                <InputLine>
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
                        validationError={validationError ?? ''}
                      />
                    )}
                  </>
                  <Actions>
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
                  </Actions>
                </InputLine>
              )}
          </>
        )
        : (
          <>
            {header
              ? (
                <Header>
                  <Typography variant="h5">
                    {value}
                  </Typography>
                </Header>
              )
              : (!!value && (
                <TextBlock>
                  <Caption
                    sx={{ color: serverError ? 'error.main' : 'text.secondary' }}
                  >
                    {!!serverError && <WarningRounded fontSize="small" />}
                    {label}
                  </Caption>
                  <Value>
                    {options.type === 'TEXT' && (
                      <Typography variant="body1">
                        {value}
                      </Typography>
                    )}

                    {options.type === 'TEXTAREA' && <TextArea value={value} tableView />}

                    {options.type === 'MULTI' && (
                      <Typography variant="body1">
                        {value}
                      </Typography>
                    )}

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
                  </Value>
                </TextBlock>
              ))}
          </>
        )}
    </>
  );
};

export default HidingCardField;
