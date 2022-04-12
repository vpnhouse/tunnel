import { AUTH } from '@constants/apiPaths';

const { API_URL } = process.env;

export const AUTH_URL = `${API_URL}${AUTH}`;
