import { createEffect, createEvent, createStore } from 'effector';

import { SETTINGS } from '@constants/apiPaths';

import { fetchData } from '../utils';
import { SettingsResponseType } from './types';

const initialSettings = null;
export const $settingsStore = createStore<SettingsResponseType | null>(initialSettings);

export const setSettings = createEvent<SettingsResponseType>();

export const getSettingsFx = createEffect<void, SettingsResponseType, Response>(
  () => fetchData(SETTINGS).then((res) => res.json())
);

export const changeSettingsFx = createEffect<SettingsResponseType, SettingsResponseType, Response>(
  (newSettings) => fetchData(
    SETTINGS,
    {
      method: 'PATCH',
      body: JSON.stringify(newSettings)
    }
  ).then((res) => res.json())
);
