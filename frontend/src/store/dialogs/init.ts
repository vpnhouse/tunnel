import { $dialogStore, closeDialog, openDialog } from '@root/store/dialogs/index';

$dialogStore
  .on(openDialog, (_, newDialog) => newDialog)
  .on(closeDialog, () => null);
