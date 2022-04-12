import { Config, ConfigFields, PasswordError } from './types';
import { PATTERN_ERRORS } from '../settings/index.constants';

export const generateRandomInterval = (min: number, max: number) => Math.floor(Math.random() * (max - min + 1)) + min;

export const generateSubMaskValue = (): string => `10.${generateRandomInterval(3, 249)}.${generateRandomInterval(3, 249)}.0/24`;

export const checkRequiredFields = (settings: Config): PasswordError => {
  const fields: ConfigFields[] = ['admin_password', 'confirm_password', 'wireguard_subnet'];

  return fields.reduce((acc, cur) => ({
    ...acc,
    ...(!settings?.[cur] ? { [cur]: PATTERN_ERRORS.required } : {})
  }), {} as PasswordError);
};
