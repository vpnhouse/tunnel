import { TRUSTED } from '@constants/apiPaths';
import { TrustedKeyRecordType } from '@root/store/trustedKeys/types';

export const TRUSTED_URL = TRUSTED;

export const EMPTY_TRUSTED_KEY: TrustedKeyRecordType = {
  id: '',
  key: ''
};
