import { AUTH_TOKEN } from '@constants/localStorageKeys';
import { logout } from '@root/store/auth';

import { CustomTimerType } from './types';

export const getAuthToken = () => localStorage.getItem(AUTH_TOKEN);

export const getTokenLifeTime = () => {
  const authToken = getAuthToken();
  if (!authToken) return 0;

  try {
    const { exp } = JSON.parse(
      atob(authToken.split('.')[1] || '')
    );
    /** time before token expires with 30 seconds reserve */
    const lifeTime = exp * 1000 - Date.now() - 30000;

    if (lifeTime < 0) throw new Error('Negative token life time');

    return lifeTime;
  } catch (e) {
    logout();
    return 0;
  }
};

export const fetchData = (
  input: Request | string,
  init?: RequestInit,
  useAuthHeader: boolean = true
): Promise<Response> => {
  const params: RequestInit = init
    ? {
      ...init,
      headers: init.headers || new Headers()
    }
    : {
      headers: new Headers()
    };

  if (useAuthHeader && params.headers instanceof Headers) {
    params.headers.set('Authorization', `Bearer ${getAuthToken()}`);
  }

  return fetch(input, params)
    .then((res) => {
      if (res.status === 401) {
        logout();
      }

      if (!res.ok) {
        return Promise.reject(res);
      }

      return res;
    });
};

export const PromiseSetTimeout = (ms: number): CustomTimerType => {
  let timeout: NodeJS.Timeout;

  const promise = new Promise((resolve) => {
    timeout = setTimeout(() => {
      resolve('timeout resolved');
    }, ms);
  });

  return {
    promise,
    clear: () => clearTimeout(timeout)
  };
};
