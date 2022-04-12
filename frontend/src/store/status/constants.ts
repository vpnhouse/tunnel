import { STATUS } from '@constants/apiPaths';

const { API_URL } = process.env;

export const STATUS_URL = `${API_URL}${STATUS}`;

export const TIMEOUT_ERROR = {
  prefix: 'timeoutError',
  message: 'Something went wrong. Service was not reloaded'
};
