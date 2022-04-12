import { DnsType } from '../index.types';

export type PropsType = {
  dns: DnsType[],
  changeDnsHandler: (dns: DnsType[]) => void;
}

export type DnsDataType = {
  dns: string;
  error: string;
}
