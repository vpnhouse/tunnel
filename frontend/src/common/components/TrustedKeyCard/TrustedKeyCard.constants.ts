import { TrustedKeyErrorType } from '@root/store/trustedKeys/types';

import { patternRequiredValidation, requiredValidation } from './TrustedKeyCard.utils';
import {
  PatternErrorType,
  TrustedKeysPatternsType,
  TrustedKeysValidationType
} from './TrustedKeyCard.types';

export const INVALID_SYMBOLS: TrustedKeysPatternsType = {
  id: /[^0-9a-fA-F-]/
};

export const SYMBOL_ERRORS: TrustedKeyErrorType = {
  id: 'Only hexadecimal digits and symbol - are allowed'
};

export const PATTERNS: TrustedKeysPatternsType = {
  id: /^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$/
};

export const PATTERN_ERRORS: PatternErrorType = {
  required: 'This field is required',
  id: 'Invalid UUID format'
};

export const PATTERN_VALIDATION: TrustedKeysValidationType = {
  id: patternRequiredValidation,
  key: requiredValidation
};
