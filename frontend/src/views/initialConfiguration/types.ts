import { Mode, ProxySchema } from '@common/components/DomainConfiguration/types';

export type Config = {
  admin_password: string;
  confirm_password: string;
  domain_name: string;
  wireguard_subnet: string;
  mode: Mode;
  issue_ssl: boolean;
  schema: ProxySchema;
}

export type ConfigFields = keyof Config;

export type ConfigTargetType = EventTarget & HTMLInputElement & {
  name: ConfigFields;
};

export type PasswordError = {
  admin_password: string;
  confirm_password: string;
  domain_name: string;
  wireguard_subnet: string;
}
