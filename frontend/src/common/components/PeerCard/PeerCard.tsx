import React, { ChangeEvent, FC, useCallback, useState, useMemo, MouseEvent, useEffect } from 'react';
import { Dialog, DialogContent, Paper, Typography } from '@material-ui/core';
import { WarningRounded } from '@material-ui/icons';
import { isFuture, isToday } from 'date-fns';
import clsx from 'clsx';

import { Button, CardField, IconButton } from '@common/ui-kit/components';
import {
  savePeerFx,
  deletePeerFx,
  changePeerFx,
  cancelCreatePeer,
  setIsEditing
} from '@root/store/peers';
import { FlatPeerType, PeerErrorType } from '@root/store/peers/types';
import { openDialog } from '@root/store/dialogs';
import DeleteIcon from '@common/assets/DeleteIcon';
import CloseIcon from '@common/assets/CloseIcon';
import { HintAdornment } from '@common/components';
import SlideTransition from '@common/components/SlideTransition';

import { PeerCardEventTargetType, PropsType } from './PeerCard.types';
import {
  INVALID_SYMBOLS,
  SYMBOL_ERRORS,
  PATTERN_VALIDATION,
  PEER_FIELD_CAN_BE_NULL,
  FAQ_CREATE_PEER_IPV4
} from './PeerCard.constants';
import { combineDateAndTime } from './PeerCard.utils';
import useStyles from './PeerCard.styles';

const PeerCard: FC<PropsType> = ({
  peerInfo,
  serverError,
  isModal,
  open,
  onClose
}) => {
  const classes = useStyles();

  const isServerError = useMemo(() => {
    if (!serverError) return false;

    return Object.values(serverError).some((error) => !!error);
  }, [serverError]);

  const [peer, setPeer] = useState<FlatPeerType>(peerInfo);
  /** There may not be server error */
  const [validationError, setValidationError] = useState<PeerErrorType | undefined>(serverError);
  const [isExpiresValid, setIsExpiresValid] = useState(true);
  const [expiresError, setExpiresError] = useState('');

  useEffect(() => {
    const keys = window?.wireguard.generateKeypair();
    setPeer(() => ({
      ...peerInfo,
      info_wireguard: {
        ...peerInfo?.info_wireguard,
        public_key: peerInfo?.info_wireguard?.public_key || keys.publicKey
      },
      private_key: peerInfo?.private_key || keys.privateKey,
      label: peerInfo?.label || 'Device name'
    }));
  }, [peerInfo]);

  useEffect(() => {
    if (!open && isModal) {
      setValidationError(serverError);
      setPeer(peerInfo);
    }
  }, [open, isModal, serverError, peerInfo]);

  const deletePeerAction = useCallback(() => {
    deletePeerFx(peer);
  }, [peer]);

  const deletePeerHandler = useCallback(() => {
    openDialog({
      title: 'Confirm peer deleting',
      message: `Are you sure you want to delete peer${peer.label ? ` ${peer.label}` : ''}?`,
      successButtonTitle: 'Delete',
      successButtonHandler: deletePeerAction
    });
  }, [peer?.label, deletePeerAction]);

  const changePeerSettings = useCallback((e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target as PeerCardEventTargetType;

    const regexp = INVALID_SYMBOLS[name];
    /** Fields without symbols patterns are not validated */
    const isInvalid = !!regexp && regexp.test(value);

    setValidationError((prevError) => ({
      ...prevError,
      [name]: isInvalid ? SYMBOL_ERRORS[name] : ''
    }));
    setPeer((prevPeer: FlatPeerType) => ({
      ...prevPeer,
      [name]: isInvalid ? prevPeer[name] : value
    }));
  }, []);

  const changeExpiresHandler = useCallback((date: Date | null, time: Date | null) => {
    /** Expires date is valid if it is empty or both date and time valid */
    const dateTimeIsEmpty = !date && !time;
    const dateIsValid = !!date && date.toString() !== 'Invalid Date' && (isToday(date) || isFuture(date));
    const timeIsValid = !!time && time.toString() !== 'Invalid Date';
    const isValid = dateTimeIsEmpty || (dateIsValid && timeIsValid);

    /** If expires date is valid set it in peer */
    isValid && setPeer((prevPeer: FlatPeerType) => ({
      ...prevPeer,
      expires: dateTimeIsEmpty
        ? null
        : date && time && combineDateAndTime(date, time).toISOString().replace('.000', '')
    }));

    /** if one field is empty and other is valid, set error */
    setExpiresError(
      (!date && timeIsValid) || (!time && dateIsValid)
        ? 'Both date and time should be either valid or empty'
        : ''
    );

    setIsExpiresValid(isValid);
  }, []);

  const cancelChangesHandler = useCallback(() => {
    if (!peer.id) {
      cancelCreatePeer();
      if (onClose) {
        onClose();
      }
      return;
    }

    setValidationError(serverError);
    setPeer(peerInfo);
    setIsEditing({
      id: peer.id,
      isEditing: false
    });
  }, [peerInfo, peer?.id, serverError, onClose]);

  const validateFormat = useCallback(() => {
    const validationErrors = (Object.entries(peer) as Entries<FlatPeerType>)
      .reduce<PeerErrorType>((errorList, [field, value]) => ({
        ...errorList,
        [field]: PATTERN_VALIDATION[field] ? (PATTERN_VALIDATION[field]?.(field, value as string) || '') : ''
      }), {} as PeerErrorType);

    const isAllFieldsValid = Object.values(validationErrors).every((error) => !error);

    if (!isAllFieldsValid || !isExpiresValid) {
      setValidationError(() => ({
        ...validationErrors,
        expires: expiresError
      }));
    }

    return isExpiresValid && isAllFieldsValid;
  }, [peer, isExpiresValid, expiresError]);

  const savePeerHandler = useCallback((event: MouseEvent) => {
    event.stopPropagation();
    event.preventDefault();

    /** If data is invalid show errors */
    if (!validateFormat()) return;

    /** Replace empty values with null */
    const nulledPeer = PEER_FIELD_CAN_BE_NULL.reduce(
      (acc, field) => ({
        ...acc,
        [field]: peer[field] || null
      }), peer
    );

    /** If peer is saved, change it, otherwise save it */
    peer.created ? changePeerFx(nulledPeer) : savePeerFx(nulledPeer);

    setIsEditing({
      id: peer.id,
      isEditing: false
    });

    if (onClose) {
      onClose();
    }
  }, [validateFormat, peer, onClose]);

  function renderForm() {
    return (
      <form className={classes.form}>
        <CardField
          isEditing={!!isModal}
          label="Device name"
          name="label"
          value={peer?.label || ''}
          validationError={validationError?.label || ''}
          serverError={serverError?.label || ''}
          options={{
            type: 'TEXT',
            textprops: {
              onChange: changePeerSettings
            }
          }}
        />

        <CardField
          isEditing={!!isModal}
          label="IPV4"
          name="ipv4"
          value={peer?.ipv4 || ''}
          validationError={validationError?.ipv4 || ''}
          serverError={serverError?.ipv4 || ''}
          isDisable
          disableControl
          options={{
            type: 'TEXT',
            textprops: {
              onChange: changePeerSettings,
              endAdornment: (<HintAdornment text={FAQ_CREATE_PEER_IPV4} />)
            }
          }}
        />

        <CardField
          isEditing={!!isModal}
          label="Expires"
          name="expires"
          value={peer?.expires || ''}
          validationError={validationError?.expires || ''}
          serverError={serverError?.expires || ''}
          readonly={!isModal}
          options={{
            type: 'DATETIME',
            onChangeHandler: changeExpiresHandler,
            dateLabel: 'Expiring Date',
            dateName: 'date',
            timeLabel: 'Expiring Time',
            timeName: 'time',
            datePickerProps: {
              disablePast: true
            }
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
            Peer is not saved! {serverError?.common}
          </Typography>
        )}
      </form>
    );
  }

  return (
    <>
      {isModal ? (
        <Dialog
          open={!!open}
          TransitionComponent={SlideTransition}
          onClose={onClose}
        >
          <DialogContent className={classes.dialog}>
            <Typography variant="h2" className={classes.title}>Create peer</Typography>

            <IconButton
              className={classes.closeDialog}
              // @ts-ignore
              onClick={onClose}
              icon={CloseIcon}
              color="primary"
            />

            {renderForm()}

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
                onClick={savePeerHandler}
              >
                Save
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      ) : (
        <Paper className={clsx(classes.paper, isServerError && classes.paperError)} elevation={0}>
          <div id="peerCardActions" className={classes.actions} onClick={deletePeerHandler}>
            <div className={classes.deleteIcon}>
              <DeleteIcon width={15} height={16} />
            </div>
          </div>

          {renderForm()}
        </Paper>
      )}
    </>
  );
};

export default PeerCard;
