import { FC, useCallback } from 'react';
import { Grow } from '@mui/material';

import { NotificationType } from '@root/store/notifications/types';
import { Alert } from '@common/ui-kit/components';
import { removeNotification } from '@root/store/notifications';

const Notification: FC<NotificationType> = ({ type, id, message, action }) => {
  const handleClose = useCallback(() =>
    removeNotification(id), [id]);

  return (
    <Grow in>
      <Alert
        message={message}
        action={action}
        severity={type}
        onClose={handleClose}
      />
    </Grow>
  );
};

export default Notification;
