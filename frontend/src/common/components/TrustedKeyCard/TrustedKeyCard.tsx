import React, { ChangeEvent, FC, useCallback, useState, useMemo, FormEvent } from 'react';
import { Paper, Typography } from '@material-ui/core';
import { Delete, Edit, WarningRounded } from '@material-ui/icons';

import { Button, CardField, IconButton } from '@common/ui-kit/components';
import { TrustedKeyErrorType, TrustedKeyRecordType } from '@root/store/trustedKeys/types';
import {
  cancelCreateTrustedKey,
  changeTrustedKeyFx,
  deleteTrustedKeyFx,
  saveTrustedKeyFx,
  setIsEditing
} from '@root/store/trustedKeys';
import { EMPTY_TRUSTED_KEY } from '@root/store/trustedKeys/constants';
import { openDialog } from '@root/store/dialogs';

import { PropsType, TrustedKeysEventTargetType } from './TrustedKeyCard.types';
import {
  INVALID_SYMBOLS,
  PATTERN_VALIDATION,
  SYMBOL_ERRORS
} from './TrustedKeyCard.constants';
import useStyles from './TrustedKeyCard.styles';

const TrustedKeyCard: FC<PropsType> = ({
  trustedKeyInfo,
  serverError,
  isEditing,
  isNotSaved = false
}) => {
  const classes = useStyles();

  const isServerError = useMemo(() => {
    if (!serverError) return false;

    return Object.values(serverError).some((error) => !!error);
  }, [serverError]);

  const [trustedKey, setTrustedKey] = useState<TrustedKeyRecordType>(trustedKeyInfo);
  /** There may not be server error */
  const [validationError, setValidationError] = useState<TrustedKeyErrorType | undefined>(serverError);

  const changeKeySettings = useCallback((e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target as TrustedKeysEventTargetType;

    const regexp = INVALID_SYMBOLS[name];
    /** Fields without symbols patterns are not validated */
    const isInvalid = !!regexp && regexp.test(value);

    setValidationError((prevError) => ({
      ...prevError,
      [name]: isInvalid ? SYMBOL_ERRORS[name] : ''
    }));
    setTrustedKey((prevKey) => ({
      ...prevKey,
      [name]: isInvalid ? prevKey[name] : value
    }));
  }, []);

  const loadKeyFromFile = useCallback((name: string, value: string) => {
    setTrustedKey((prevKey) => ({
      ...prevKey,
      [name]: value
    }));
  }, []);

  const editKeyHandler = useCallback(() => {
    setIsEditing({
      id: trustedKey.id,
      isEditing: true
    });
  }, [trustedKey.id]);

  const cancelChangesHandler = useCallback(() => {
    if (!trustedKeyInfo.id) {
      cancelCreateTrustedKey();
      return;
    }

    setValidationError(serverError);
    setTrustedKey(trustedKeyInfo);
    setIsEditing({
      id: trustedKey.id,
      isEditing: false
    });
  }, [serverError, trustedKeyInfo, trustedKey.id]);

  const deleteKeyAction = useCallback(() => {
    deleteTrustedKeyFx({
      id: trustedKeyInfo.id,
      isNotSaved
    });
  }, [trustedKeyInfo, isNotSaved]);

  const deleteKeyHandler = useCallback(() => {
    openDialog({
      title: 'Confirm trusted key deleting',
      message: 'Are you sure you want to delete trusted key?',
      successButtonTitle: 'Delete',
      successButtonHandler: deleteKeyAction
    });
  }, [deleteKeyAction]);

  const validateFormat = useCallback(() => {
    const validationErrors = (Object.entries(trustedKey) as Entries<TrustedKeyRecordType>)
      .reduce<TrustedKeyErrorType>((errorList, [field, value]) => ({
        ...errorList,
        [field]: PATTERN_VALIDATION[field] ? PATTERN_VALIDATION[field](field, value) : ''
      }), EMPTY_TRUSTED_KEY);
    const isAllFieldsValid = Object.values(validationErrors).every((error) => !error);

    if (!isAllFieldsValid) setValidationError(validationErrors);

    return isAllFieldsValid;
  }, [trustedKey]);

  const saveKeyHandler = useCallback((event: FormEvent<HTMLFormElement>) => {
    event.stopPropagation();
    event.preventDefault();

    /** If data is invalid show errors */
    if (!validateFormat()) return;

    isNotSaved
      ? saveTrustedKeyFx({
        trustedKeyInfo: trustedKey,
        isNotSaved,
        prevId: trustedKeyInfo.id
      })
      : changeTrustedKeyFx(trustedKey);
    setIsEditing({
      id: trustedKey.id,
      isEditing: false
    });
  }, [validateFormat, trustedKey, isNotSaved, trustedKeyInfo]);

  return (
    <Paper className={isServerError ? `${classes.paper} ${classes.paperError}` : classes.paper} elevation={0}>
      {!isEditing && (
        <div className={classes.actions}>
          <IconButton
            color="primary"
            onClick={editKeyHandler}
            icon={Edit}
          />
          <IconButton
            color="error"
            onClick={deleteKeyHandler}
            icon={Delete}
          />
        </div>
      )}

      <form onSubmit={saveKeyHandler} className={classes.form}>
        <CardField
          isEditing={isEditing && isNotSaved}
          label="UUID"
          name="id"
          value={trustedKey?.id || ''}
          validationError={validationError?.id || ''}
          serverError={serverError?.id || ''}
          options={{
            type: 'TEXT',
            textprops: {
              onChange: changeKeySettings
            }
          }}
        />

        <CardField
          isEditing={isEditing}
          label="Public Key"
          name="key"
          value={trustedKey?.key || ''}
          validationError={validationError?.key || ''}
          serverError={serverError?.key || ''}
          copyToClipboard
          options={{
            type: 'TEXTAREA',
            textprops: {
              onChange: changeKeySettings,
              rows: 6,
              className: classes.publicKey
            }
          }}
          loadOptions={{
            accept: '.pem',
            onLoad: loadKeyFromFile
          }}
        />

        {!!serverError?.common && (
          <Typography
            className={classes.commonError}
            variant="caption"
            component="div"
            color="error"
          >
            <WarningRounded fontSize="small" />
            Key is not saved! {serverError?.common}
          </Typography>
        )}

        {isEditing && (
          <div className={classes.buttonLine}>
            <Button
              variant="contained"
              color="secondary"
              onClick={cancelChangesHandler}
            >
              Cancel
            </Button>
            <Button
              variant="contained"
              color="primary"
              type="submit"
            >
              Save changes
            </Button>
          </div>
        )}
      </form>
    </Paper>
  );
};

export default TrustedKeyCard;
