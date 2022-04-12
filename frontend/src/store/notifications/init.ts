import {
  $notificationsStore,
  createNotification,
  startRemoveTimerFx,
  removeAllNotifications,
  removeNotification
} from './index';

$notificationsStore
  .on(createNotification, (store, notification) => [
    ...store,
    notification
  ])
  .on(removeNotification, (store, id) => {
    const item = store.find(({ notification }) => notification.id === id);
    if (item?.timer) item.timer.clear();

    return store.filter(({ notification }) => notification.id !== id);
  });

/** If notification has timer to autoremove, start it */
createNotification.watch(
  (notification) => notification?.timer && startRemoveTimerFx(notification)
);

startRemoveTimerFx.done.watch(({ params }) =>
  removeNotification(params.notification.id));

$notificationsStore.reset(removeAllNotifications);
