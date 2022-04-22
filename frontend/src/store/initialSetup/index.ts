import { createEffect, createEvent, createStore } from 'effector';

import { AUTH, INITIAL_SETUP } from '@constants/apiPaths';

import { fetchData } from '../utils';
import { InitialSetupData } from './types';

export const $initialSetup = createStore(false);

export const setInitialSetupState = createEvent<boolean>();

export const checkConfigurationFx = createEffect<void, void, Response>(
  () => fetchData(AUTH, { method: 'GET' }).then((res) => res.json())
);

export const setInitialSetupFx = createEffect<InitialSetupData, InitialSetupData, Response>(
  (data: InitialSetupData) => fetchData(
    INITIAL_SETUP,
    {
      method: 'POST',
      body: JSON.stringify(data)
    }
  ).then(() => data)
);
