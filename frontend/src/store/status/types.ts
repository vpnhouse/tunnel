import { components } from '@schema';

export type StatusResponseType = components['schemas']['ServiceStatus'];

export type StatusTimerType = {
  intervalTimer: NodeJS.Timeout | null;
  stopTimer: NodeJS.Timeout | null;
}

export type StopStatusTimerModeType = 'timeout' | 'restart';

export type StopStatusTimerType = {
 timer: StatusTimerType;
 mode: StopStatusTimerModeType;
}
