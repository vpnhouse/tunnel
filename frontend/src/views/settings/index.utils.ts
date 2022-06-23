import { ip4ToInt, isSubnetPrivate, maskFromBits } from '@root/common/utils/ipv4';

import { PATTERN_ERRORS, PATTERNS } from './index.constants';

export const addIdtoDns = (dns: string[]) =>
  dns.map((item) => ({
    id: item,
    dns: item,
    error: ''
  }));

export const dnsNameValidation = (value: string) => {
  if (value && !PATTERNS.dnsName.test(value)) return PATTERN_ERRORS.dnsName;

  return '';
};

export const ipv4Validation = (value: string) => {
  if (value && !PATTERNS.ipv4.test(value)) return PATTERN_ERRORS.ipv4;

  return '';
};

export const portValidation = (value: number) => {
  if (value < 1 || value > 65535) return PATTERN_ERRORS.port;

  return '';
};

export const subnetValidation = (value: string) => {
  if (!value) return PATTERN_ERRORS.required;
  if (!PATTERNS.cidr.test(value)) return PATTERN_ERRORS.cidr;

  const [subnetString, bitsString] = value.split('/');

  const subnetInteger = ip4ToInt(subnetString);
  const mask = maskFromBits(+bitsString);
  const invertedMask = ~mask;

  if ((subnetInteger & invertedMask) !== 0) {
    return PATTERN_ERRORS.ipToSubnet;
  }

  if (!isSubnetPrivate(subnetInteger)) {
    return PATTERN_ERRORS.ipToSubnet;
  }

  return '';
};

export const dnsValidation = (value: string) => {
  if (value && !PATTERNS.ipv4.test(value)) return PATTERN_ERRORS.ipv4;

  return '';
};
