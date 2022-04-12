import { ConfigFields } from '@root/views/initialConfiguration/types';
import { SettingsType } from '@root/store/settings/types';
import { DomainConfig } from '@common/components/DomainConfiguration/types';

import { INVALID_SYMBOLS } from './index.constants';

export type DnsType = {
  dns: string;
  id: string;
  error: string;
};

export type SymbolsPatternsNamesType = keyof typeof INVALID_SYMBOLS;

export type SettingsNumericFields = keyof Pick<SettingsType, 'connection_timeout' | 'ping_interval' | 'wireguard_keepalive' | 'wireguard_listen_port'
  | 'wireguard_server_port'>;
export type SettingsStringFields = keyof Pick<SettingsType, 'wireguard_public_key' | 'wireguard_server_ipv4' | 'dns'| 'wireguard_subnet' | 'admin_password' | 'confirm_password'>
  | keyof Pick<DomainConfig, 'domain_name'>;
export type SettingsInputNamesType = SettingsNumericFields | SettingsStringFields;

export type SymbolsSchemesType = {
  [key in ConfigFields | SettingsInputNamesType]: SymbolsPatternsNamesType;
};

export type SettingsKeysType = keyof Omit<SettingsType, 'wireguard_public_key'>;

export type SettingsChangedType = {
  [key in SettingsKeysType]: boolean;
} & {
  domain: boolean;
};

export type SettingsErrorType = {
  [key in SettingsKeysType]: string;
} & {
  domain_name: string
};

export type SettingsEventTargetType = EventTarget & HTMLInputElement & {
  name: SettingsKeysType;
};

export type PatternValidationType = {
  [key in SettingsNumericFields]?: (value: number) => string;
} & {
  [key in SettingsStringFields]?: (value: string) => string;
};
