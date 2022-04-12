import { ReactNode } from 'react';

import { CustomTimerType } from '../types';

export type NotificationTemplateType = {
  type: 'error' | 'info' | 'warning',
  /** Prefix for notification id */
  prefix: string,
  message: string,
  /** Action (like button) that can be added to the notification instead of default close button */
  action?: ReactNode
}

export type NotificationType = {
  type: 'error' | 'info' | 'warning',
  /** Unique id: prefix plus unique number */
  id: string,
  message: string,
  /** Action (like button) that can be added to the notification instead of default close button */
  action?: ReactNode
}

export type NotificationContainer = {
  timer?: CustomTimerType,
  notification: NotificationType
}
