import { createStore, createEvent } from 'effector';

import { DialogType } from './types';

export const $dialogStore = createStore<DialogType | null>(null);
export const openDialog = createEvent<DialogType>();
export const closeDialog = createEvent();
