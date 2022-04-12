import { AUTH_TOKEN } from '@constants/localStorageKeys';
import { AUTH_ERROR, REFRESH_TOKEN_ERROR } from '@constants/notifications';

import {
  $authStore,
  checkToken,
  setToken,
  logout,
  $tokenTimerStore,
  loginFx,
  refreshTokenFx,
  refreshToken,
  setTimer
} from './index';
import { addNotification, removeAllNotifications, showServerErrorFx } from '../notifications';
import { getAllPeersFx } from '../peers';
import { getTokenLifeTime } from '../utils';

$authStore
  .on(checkToken, () => !!getTokenLifeTime())
  .on(setToken, (_, token) => {
    localStorage.setItem(AUTH_TOKEN, token);

    return true;
  })
  .on(logout, () => {
    localStorage.removeItem(AUTH_TOKEN);

    return false;
  });

$tokenTimerStore
  .on(refreshToken, (_, tokenTimer) => tokenTimer)
  .on(logout, (tokenTimer) => {
    if (tokenTimer) tokenTimer.clear();

    return null;
  });


/** Set token to localStorage after authentication and remove all notifications */
loginFx.doneData.watch((result) => {
  setToken(result.access_token);
  removeAllNotifications();
});

/** Handle error if authentication data is invalid */
loginFx.failData.watch((error) => {
  if (error.status === 401) {
    addNotification({
      type: 'error',
      ...AUTH_ERROR
    });

    return;
  }

  showServerErrorFx(error);
});

/** If refreshing is successful, set new token to localStorage */
refreshTokenFx.doneData.watch((result) => {
  setToken(result.access_token);
});

/** if refreshing failed logout and show error */
refreshTokenFx.failData.watch(() => {
  logout();
  addNotification({
    type: 'error',
    ...REFRESH_TOKEN_ERROR
  });
});

refreshToken.watch((tokenTimer) => {
  if (tokenTimer) refreshTokenFx(tokenTimer);
});

/** Set timer for token refresh after checking token on first load
 * and every time token changes */
checkToken.watch(() => setTimer());
setToken.watch(() => setTimer());

$authStore.watch((state) => {
  if (state) getAllPeersFx();
});
