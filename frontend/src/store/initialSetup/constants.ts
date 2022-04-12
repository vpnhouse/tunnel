import { INITIAL_SETUP } from '@root/constants/apiPaths';

const { API_URL } = process.env;

export const INITIAL_SETUP_URL = `${API_URL}${INITIAL_SETUP}`;
