import { useEffect } from 'react';
import { useUnit } from 'effector-react';
import { useLocation } from 'react-router';

import { $notificationsStore, removeAllNotifications } from '@root/store/notifications';
import { $authStore } from '@root/store/auth';

import Notification from './Notification/Notification';
import useStyles from './NotificationsBar.styles';

const NotificationsBar = () => {
  const { pathname } = useLocation();
  const notifications = useUnit($notificationsStore).slice(-3);
  const isAuthenticated = useUnit($authStore);
  const classes = useStyles();

  useEffect(() => {
    removeAllNotifications();
  }, [pathname]);

  return (
    <div className={`${classes.stack} ${isAuthenticated ? classes.authShift : ''}`}>
      {notifications.map(({ notification }) => (
        <Notification
          key={notification.id}
          {...notification}
        />
      ))}
    </div>
  );
};

export default NotificationsBar;
