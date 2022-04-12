import {
  dnsNameValidation,
  dnsValidation,
  ipv4Validation,
  portValidation,
  subnetValidation
} from './index.utils';
import { PatternValidationType, SymbolsSchemesType } from './index.types';

export const NUMERIC_FIELDS = ['wireguard_keepalive', 'wireguard_server_port', 'connection_timeout', 'ping_interval', 'wireguard_listen_port'];

export const INVALID_SYMBOLS = {
  printableASCIINotSpace: /[^\x21-\x7E]+/,
  printableASCII: /[^\x20-\x7E]+/,
  dnsName: /[^a-zA-Z0-9._-]/,
  ipv4: /[^0-9.]/,
  dns: /[^0-9.]/,
  numbers: /[^0-9]/,
  cidr: /[^0-9./]/
};

export const SYMBOL_ERRORS = {
  printableASCIINotSpace: 'The field may contain only printable ASCII symbols without spaces',
  printableASCII: 'The field may contain only printable ASCII symbols',
  dnsName: 'DNS name may contain only letters, digits and symbols ._-',
  ipv4: 'IPV4 address may contain only digits and dots',
  dns: 'DNS may contain only digits and dots',
  numbers: 'The field may contain only positive whole numbers',
  cidr: 'Subnet mask may contain only digits and symbols ./'
};

export const SYMBOL_SCHEMES: SymbolsSchemesType = {
  admin_password: 'printableASCII',
  domain_name: 'dnsName',
  wireguard_keepalive: 'numbers',
  wireguard_server_ipv4: 'ipv4',
  wireguard_server_port: 'numbers',
  wireguard_subnet: 'cidr',
  dns: 'dns',
  connection_timeout: 'numbers',
  confirm_password: 'printableASCII',
  schema: 'printableASCII',
  mode: 'printableASCII',
  ping_interval: 'numbers',
  issue_ssl: 'printableASCII',
  wireguard_listen_port: 'numbers',
  wireguard_public_key: 'printableASCII'
};

export const PATTERNS = {
  dnsName: /^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9](?:\.[a-zA-Z]{2,})+$/,
  ipv4: /^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/,
  cidr: /^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\/([0-2]?[0-9]|[3][0-2]?)$/
};

export const PATTERN_ERRORS = {
  dnsName: 'Invalid DNS name format',
  ipv4: 'Invalid IPV4 format',
  cidr: 'Invalid subnet mask format. It must match pattern: IPV4/number between 0 and 32',
  port: 'Port number must be a whole number between 1 and 65535',
  required: 'This field is required',
  password: 'Passwords need to match'
};

export const PATTERN_VALIDATION: PatternValidationType = {
  domain_name: dnsNameValidation,
  wireguard_server_ipv4: ipv4Validation,
  wireguard_server_port: portValidation,
  wireguard_subnet: subnetValidation,
  dns: dnsValidation
};
