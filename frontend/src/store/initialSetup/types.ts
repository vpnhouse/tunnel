import { Mode, ProxySchema } from '@common/components/DomainConfiguration/types';

export interface InitialSetupDomain {
  mode: Mode;
  domain_name: string;
  issue_ssl?: boolean;
  schema?: ProxySchema;
}

export type InitialSetupData = {
  admin_password: string;
  server_ip_mask: string;
  send_stats: boolean;
  domain?: InitialSetupDomain;
}
