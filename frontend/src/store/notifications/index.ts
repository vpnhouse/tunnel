import { createStore, createEffect, createEvent } from 'effector';

import { NotificationContainer, NotificationTemplateType } from './types';
import { INFO_FADE_MS } from './constants';
import { PromiseSetTimeout } from '../utils';

const initialStore: NotificationContainer[] = [];

export const $notificationsStore = createStore(initialStore);
export const createNotification = createEvent<NotificationContainer>();
export const removeNotification = createEvent<string>();
export const removeAllNotifications = createEvent();

export const addNotification = createNotification.prepend(
  (notification: NotificationTemplateType) => {
    switch (notification.type) {
      case 'info':
        return {
          notification: {
            ...notification,
            id: `${notification.prefix}-${Date.now()}`
          },
          timer: PromiseSetTimeout(INFO_FADE_MS)
        };

      default:
        return {
          notification: {
            ...notification,
            id: `${notification.prefix}-${Date.now()}`
          }
        };
    }
  }
);

export const startRemoveTimerFx = createEffect(
  ({ timer }: NotificationContainer) => timer?.promise
);

export const showServerErrorFx = createEffect((error: Response) => {
  error.json().then((errorDetails) => {
    addNotification({
      type: 'error',
      prefix: 'serverError',
      message: errorDetails.error
    });
  });
});
