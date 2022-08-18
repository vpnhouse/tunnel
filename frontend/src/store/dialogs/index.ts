import { createStore, createEvent } from 'effector';

import { DialogStore, DialogType } from './types';

export const $dialogStore = createStore<DialogStore | null>(null);
export const openDialog = createEvent<DialogType>();
export const closeDialog = createEvent();
