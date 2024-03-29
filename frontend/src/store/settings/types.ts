import { WithDomain } from '@common/components/DomainConfiguration/types';

export type SettingsResponseType = WithDomain & {
  connection_timeout: number;
  dns: string[];
  ping_interval: number;
  wireguard_keepalive: number;
  wireguard_listen_port: number;
  wireguard_public_key: string;
  wireguard_server_ipv4: string;
  wireguard_server_port: number;
  wireguard_subnet: string;
  send_stats: boolean;
};

export type SettingsType = Omit<SettingsResponseType, 'domain'> & {
  confirm_password?: string;
  admin_password?: string;
}

export type SettingsRequest = SettingsResponseType & {
  admin_password?: string;
}
