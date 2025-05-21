import { createEffect, createEvent, createStore } from 'effector';

import { STATUS } from '@constants/apiPaths';

import { fetchData } from '../utils';
import { StopStatusTimerModeType, StopStatusTimerType, StatusResponseType, StatusTimerType } from './types';

const initialStatus = {
  restart_required: false,
  stats_global: {
    peers_total: 0,
    peers_active: 0,
    traffic_up: 0,
    traffic_down: 0,
    speed_down: 0,
    speed_up: 0
  }
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
  () => fetchData(STATUS)
    .then((res) => res.json())
    .catch((err) => {
      throw err;
    })
);
