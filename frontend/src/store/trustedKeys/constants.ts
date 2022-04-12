import { TRUSTED } from '@constants/apiPaths';
import { TrustedKeyRecordType } from '@root/store/trustedKeys/types';

const { API_URL } = process.env;

export const TRUSTED_URL = `${API_URL}${TRUSTED}`;

export const EMPTY_TRUSTED_KEY: TrustedKeyRecordType = {
  id: '',
  key: ''
};
