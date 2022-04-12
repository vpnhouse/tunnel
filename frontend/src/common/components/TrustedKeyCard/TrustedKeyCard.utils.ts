import { PATTERN_ERRORS, PATTERNS } from './TrustedKeyCard.constants';
import { TrustedKeysFieldsType } from './TrustedKeyCard.types';

export const patternRequiredValidation = (field: TrustedKeysFieldsType, value: string): string => {
  if (!value) return PATTERN_ERRORS.required;

  const isValid = PATTERNS[field]?.test(value);

  return isValid ? '' : (PATTERN_ERRORS[field] || '');
};

export const requiredValidation = (field: TrustedKeysFieldsType, value: string): string =>
  (value ? '' : PATTERN_ERRORS.required);
