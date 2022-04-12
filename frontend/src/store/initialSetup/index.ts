import { createEffect, createEvent, createStore } from 'effector';

import { AUTH_URL } from '../auth/constants';
import { fetchData } from '../utils';
import { INITIAL_SETUP_URL } from './constants';
import { InitialSetupData } from './types';

export const $initialSetup = createStore(false);

export const setInitialSetupState = createEvent<boolean>();

export const checkConfigurationFx = createEffect<void, void, Response>(
  () => fetchData(AUTH_URL, { method: 'GET' }).then((res) => res.json())
);

export const setInitialSetupFx = createEffect<InitialSetupData, InitialSetupData, Response>(
  (data: InitialSetupData) => fetchData(
    INITIAL_SETUP_URL,
    {
      method: 'POST',
      body: JSON.stringify(data)
    }
  ).then(() => data)
);
