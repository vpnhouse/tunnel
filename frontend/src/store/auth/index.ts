import { createStore, createEffect, createEvent } from 'effector';

import { fetchData, getTokenLifeTime, PromiseSetTimeout } from '../utils';
import { CustomTimerType } from '../types';
import { AUTH_URL } from './constants';
import { AuthDataType, AuthResponseType } from './types';

export const $authStore = createStore(false);
export const $tokenTimerStore = createStore<CustomTimerType | null>(null);

export const checkToken = createEvent();
export const setToken = createEvent<string>();
export const logout = createEvent();

/** Authenticate and get access token */
export const loginFx = createEffect<AuthDataType, AuthResponseType, Response>(({ password }) => {
  const headers = new Headers();
  const unicodePassword = unescape(encodeURIComponent(password));

  headers.set('Authorization', `Basic ${btoa(`:${unicodePassword}`)}`);

  return fetchData(
    AUTH_URL,
    {
      headers
    },
    false
  ).then((res) => res.json());
});

/** Refresh token before it expires */
export const refreshTokenFx = createEffect<CustomTimerType, AuthResponseType, Response>(
  (tokenTimer) => tokenTimer.promise
    .then(() => fetchData(AUTH_URL).then((res) => res.json()))
);

/** Events to start token refreshing */
export const refreshToken = createEvent<null | CustomTimerType>();
export const setTimer = refreshToken.prepend(() => {
  const tokenLifeTime = getTokenLifeTime();

  return tokenLifeTime ? PromiseSetTimeout(tokenLifeTime) : null;
});
