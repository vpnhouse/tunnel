import { FC, useCallback } from 'react';
import { useUnit } from 'effector-react';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

import { $dialogStore, closeDialog } from '@root/store/dialogs';
import { Button, IconButton } from '@common/ui-kit/components';
import CloseIcon from '@common/assets/CloseIcon';
import SlideTransition from '@common/components/SlideTransition';

const StyledDialog = styled(Dialog)(({ theme }) => ({
  '& .MuiDialog-paper': {
    padding: 32,
    maxWidth: 770,
    borderRadius: 12
  }
}));

const StyledDialogTitle = styled(DialogTitle)(({ theme }) => ({
  ...theme.typography.h5,
  padding: 0,
  fontWeight: 500,
  marginBottom: 24,
  fontSize: 24,
  lineHeight: '32px'
}));

const CloseButton = styled(Box)({
  position: 'absolute',
  top: 42,
  right: 42
});

const StyledDialogContent = styled(DialogContent)({
  padding: 0
});

const StyledDialogContentText = styled(DialogContentText)(({ theme }) => ({
  ...theme.typography.subtitle1,
  color: theme.palette.text.primary,
  marginBottom: '32px'
}));

const StyledDialogActions = styled(DialogActions)({
  padding: 0,
  '& > :not(:first-of-type)': {
    marginLeft: '12px'
  }
});

const ButtonsContainer = styled(Box)({
  display: 'flex',
  justifyContent: 'space-between',
  width: '100%',
  marginTop: 24
});

const DialogComponent: FC = () => {
  const dialog = useUnit($dialogStore);

  const closeHandler = useCallback(() => {
    closeDialog();
  }, []);

  const successHandler = useCallback(() => {
    closeDialog();
    dialog?.successButtonHandler?.();
  }, [dialog]);

  if (!dialog) return null;

  return (
    <StyledDialog
      open={dialog.opened}
      TransitionComponent={SlideTransition}
      onClose={closeHandler}
    >
      <StyledDialogTitle>
        {dialog.title}
      </StyledDialogTitle>

      <CloseButton>
        <IconButton
          onClick={closeHandler}
          icon={CloseIcon}
          color="primary"
        />
      </CloseButton>

      <StyledDialogContent>
        {typeof dialog.message === 'string' ? (
          <StyledDialogContentText>
            {dialog.message}
          </StyledDialogContentText>
        ) : dialog.message}
      </StyledDialogContent>

      <StyledDialogActions>
        {dialog.actionComponent ? (
          <ButtonsContainer>
            {dialog.actionComponent}
          </ButtonsContainer>
        ) : (
          <ButtonsContainer>
            <Button onClick={closeHandler} variant="contained" color="secondary">
              Cancel
            </Button>
            <Button onClick={successHandler} variant="contained" color="primary">
              {dialog.successButtonTitle}
            </Button>
          </ButtonsContainer>
        )}
      </StyledDialogActions>
    </StyledDialog>
  );
};

export default DialogComponent;
