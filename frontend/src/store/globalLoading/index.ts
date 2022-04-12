import { createEvent, createStore } from 'effector';

export const $globalLoading = createStore(true);

export const setGlobalLoading = createEvent<boolean>();
