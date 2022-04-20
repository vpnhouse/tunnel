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
import { Button } from '@common/ui-kit/components';

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
      open={!!dialog}
      disableBackdropClick
      classes={{
        paper: classes.paper
      }}
    >
      <DialogTitle
        disableTypography
        classes={{
          root: classes.title
        }}
      >
        {dialog?.title}
      </DialogTitle>
      <DialogContent
        classes={{
          root: classes.content
        }}
      >
        {typeof dialog?.message === 'string' ? (
          <DialogContentText
            classes={{
              root: classes.contentText
            }}
          >
            {dialog?.message}
          </DialogContentText>
        ) : dialog?.message}

      </DialogContent>
      <DialogActions
        classes={{
          root: classes.actions
        }}
      >
        {dialog.onlyClose ? (
          <div className={classes.buttons}>
            <Button variant="contained" color="secondary">
              Download file
            </Button>
            <Button onClick={closeHandler} variant="contained" color="secondary">
              Close
            </Button>
          </div>
        ) : (
          <>
            <Button onClick={closeHandler} variant="contained" color="secondary">
              Cancel
            </Button>
            <Button onClick={successHandler} variant="contained" color="primary">
              {dialog?.successButtonTitle}
            </Button>
          </>
        )}
      </DialogActions>
    </MaterialDialog>
  );
};

export default Dialog;
