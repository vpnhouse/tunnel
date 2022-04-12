import { DnsType } from '../../index.types';
import { DnsDataType } from '../DnsSettings.types';

export type PropsType = {
  dns: DnsType;
  removeDnsFromList: (id: string) => void;
  onDnsChange: (id: string, dnsData: DnsDataType) => void;
  autoFocus: boolean;
  onBlur: () => void;
}
