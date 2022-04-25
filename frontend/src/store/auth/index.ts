import { createEffect, createEvent, createStore } from 'effector';
import { encode } from 'js-base64';

import { AUTH } from '@constants/apiPaths';

import { fetchData, getTokenLifeTime, PromiseSetTimeout } from '../utils';
import { CustomTimerType } from '../types';
import { AuthDataType, AuthResponseType } from './types';

export const $authStore = createStore(false);
export const $tokenTimerStore = createStore<CustomTimerType | null>(null);

export const checkToken = createEvent();
export const setToken = createEvent<string>();
export const logout = createEvent();

/** Authenticate and get access token */
export const loginFx = createEffect<AuthDataType, AuthResponseType, Response>(({ password }) => {
  const headers = new Headers();

  // 3rd party encoding function is used because btoa doesn't support non-ascii characters
  headers.set('Authorization', `Basic ${encode(`:${password}`)}`);

  return fetchData(
    AUTH,
    {
      headers
    },
    false
  ).then((res) => res.json());
});

/** Refresh token before it expires */
export const refreshTokenFx = createEffect<CustomTimerType, AuthResponseType, Response>(
  (tokenTimer) => tokenTimer.promise
    .then(() => fetchData(AUTH).then((res) => res.json()))
);

/** Events to start token refreshing */
export const refreshToken = createEvent<null | CustomTimerType>();
export const setTimer = refreshToken.prepend(() => {
  const tokenLifeTime = getTokenLifeTime();

  return tokenLifeTime ? PromiseSetTimeout(tokenLifeTime) : null;
});
