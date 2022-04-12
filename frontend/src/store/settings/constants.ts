import { SETTINGS } from '@constants/apiPaths';

const { API_URL } = process.env;

export const SETTINGS_URL = `${API_URL}${SETTINGS}`;
