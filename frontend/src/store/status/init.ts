import { sample } from 'effector';

import {
  $statusStore,
  checkStatusFx,
  setStatus,
  $loadingStore,
  setLoading,
  checkStatus,
  $statusTimerStore,
  setStatusTimer,
  clearStatusTimer,
  stopStatusTimerFx, stopStatusTimer
} from './index';
import { addNotification } from '../notifications';
import { TIMEOUT_ERROR } from './constants';

$statusStore
  .on(setStatus, (state, status) => ({
    ...state,
    ...status
  }));

$statusTimerStore
  .on(setStatusTimer, (_, timer) => timer)
  .on(clearStatusTimer, () => ({
    intervalTimer: null,
    stopTimer: null
  }));

$loadingStore
  .on(setLoading, (_, isLoading) => isLoading);

checkStatus.watch(() => {
  setLoading(true);
  const intervalId = setInterval(() => checkStatusFx(), 500);

  const stopId = setTimeout(() => stopStatusTimer('timeout'), 10000);
  setStatusTimer({
    intervalTimer: intervalId,
    stopTimer: stopId
  });
});

checkStatusFx.doneData.watch(
  (status) => {
    setStatus(status);
    !status.restart_required && stopStatusTimer('restart');
  }
);

sample({
  clock: stopStatusTimer,
  source: $statusTimerStore,
  fn: (timer, mode) => ({
    timer,
    mode
  }),
  target: stopStatusTimerFx
});

stopStatusTimerFx.doneData.watch((result) => {
  if (result) {
    addNotification({
      type: 'error',
      ...TIMEOUT_ERROR
    });
    setStatus({
      restart_required: true,
      stats_global: {
        peers_total: 0,
        peers_active: 0,
        traffic_up: 0,
        traffic_down: 0,
        speed_down: 0,
        speed_up: 0
      }
    });
    setLoading(false);
  }

  clearStatusTimer();
});

checkStatusFx.failData.watch(
  (error) => error.json().then((err) =>
    addNotification({
      type: 'error',
      prefix: 'serverError',
      message: err.error
    }))
);

$statusStore.watch((state) =>
  setLoading(state.restart_required));
