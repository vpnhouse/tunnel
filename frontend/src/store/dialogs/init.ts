import { $dialogStore, closeDialog, openDialog } from '@root/store/dialogs/index';
import { DialogStore } from '@root/store/dialogs/types';

$dialogStore
  .on(openDialog, (_, newDialog) => ({
    ...newDialog,
    opened: true
  }))
  .on(closeDialog, (store) => ({
    ...store,
    opened: false
  } as DialogStore));
