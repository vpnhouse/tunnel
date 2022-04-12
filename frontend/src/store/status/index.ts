import { createEffect, createEvent, createStore } from 'effector';

import { fetchData } from '../utils';
import { StopStatusTimerModeType, StopStatusTimerType, StatusResponseType, StatusTimerType } from './types';
import { STATUS_URL } from './constants';

const initialStatus = {
  restart_required: false
};

const initialTimer = {
  intervalTimer: null,
  stopTimer: null
};

export const $statusStore = createStore<StatusResponseType>(initialStatus);
export const setStatus = createEvent<StatusResponseType>();
export const checkStatus = createEvent();

export const $statusTimerStore = createStore<StatusTimerType>(initialTimer);
export const setStatusTimer = createEvent<StatusTimerType>();
export const clearStatusTimer = createEvent();
export const stopStatusTimer = createEvent<StopStatusTimerModeType>();
export const stopStatusTimerFx = createEffect<StopStatusTimerType, boolean>(({ timer, mode }) => {
  const { intervalTimer, stopTimer } = timer;
  intervalTimer && clearInterval(intervalTimer);
  stopTimer && clearInterval(stopTimer);

  return !!intervalTimer && (mode === 'timeout');
});

export const $loadingStore = createStore(false);
export const setLoading = createEvent<boolean>();

export const checkStatusFx = createEffect<void, StatusResponseType, Response>(
  () => fetchData(STATUS_URL)
    .then((res) => res.json())
    .catch((err) => {
      throw err;
    })
);
