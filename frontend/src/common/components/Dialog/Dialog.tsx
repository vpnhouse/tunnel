import React, { FC, useCallback } from 'react';
import { useStore } from 'effector-react';
import {
  Dialog as MaterialDialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle
} from '@material-ui/core';

import { $dialogStore, closeDialog } from '@root/store/dialogs';
import { Button, IconButton } from '@common/ui-kit/components';
import CloseIcon from '@common/assets/CloseIcon';
import SlideTransition from '@common/components/SlideTransition';

import useStyles from './Dialog.styles';

const Dialog: FC = () => {
  const dialog = useStore($dialogStore);
  const classes = useStyles();

  const closeHandler = useCallback(() => {
    closeDialog();
  }, []);

  const successHandler = useCallback(() => {
    closeDialog();
    dialog?.successButtonHandler && dialog.successButtonHandler();
  }, [dialog]);

  if (!dialog) return null;

  return (
    <MaterialDialog
      open={dialog.opened}
      TransitionComponent={SlideTransition}
      classes={{
        paper: classes.paper
      }}
      onClose={closeHandler}
    >
      <DialogTitle
        disableTypography
        classes={{
          root: classes.title
        }}
      >
        {dialog.title}
      </DialogTitle>

      <IconButton
        className={classes.closeDialog}
        // @ts-ignore
        onClick={closeHandler}
        icon={CloseIcon}
        color="primary"
      />

      <DialogContent
        classes={{
          root: classes.content
        }}
      >
        {typeof dialog.message === 'string' ? (
          <DialogContentText
            classes={{
              root: classes.contentText
            }}
          >
            {dialog.message}
          </DialogContentText>
        ) : dialog.message}

      </DialogContent>
      <DialogActions
        classes={{
          root: classes.actions
        }}
      >
        {dialog.actionComponent ? (
          <div className={classes.buttons}>
            {dialog.actionComponent}
          </div>
        ) : (
          <div className={classes.buttons}>
            <Button onClick={closeHandler} variant="contained" color="secondary">
              Cancel
            </Button>
            <Button onClick={successHandler} variant="contained" color="primary">
              {dialog.successButtonTitle}
            </Button>
          </div>
        )}
      </DialogActions>
    </MaterialDialog>
  );
};

export default Dialog;
