import { PATTERN_ERRORS, PATTERNS } from '@common/components/PeerCard/PeerCard.constants';

export const patternValidation = (field: string, value: string): string => {
  const isValid = !value || PATTERNS[field]?.test(value);

  return isValid ? '' : (PATTERN_ERRORS[field] || '');
};

export const patternRequiredValidation = (field: string, value: string): string => {
  if (!value) return PATTERN_ERRORS.required;

  const isValid = PATTERNS[field]?.test(value);

  return isValid ? '' : (PATTERN_ERRORS[field] || '');
};

export const combineDateAndTime = (date: Date, time: Date): Date => {
  const year = date.getFullYear();
  const month = date.getMonth();
  const day = date.getDate();
  const hours = time.getHours();
  const minutes = time.getMinutes();
  const seconds = time.getSeconds();

  return new Date(year, month, day, hours, minutes, seconds);
};
