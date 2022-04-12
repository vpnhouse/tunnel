export enum ProxySchema {
  http = 'http',
  https = 'https'
}

export enum Mode {
  Direct = 'direct',
  ReverseProxy = 'reverse-proxy'
}

export type DomainConfig = {
  domain_name: string;
  mode: Mode;
  schema?: ProxySchema;
  issue_ssl?: boolean;
}

export type WithDomain = {
  domain: DomainConfig | null;
}

export type DomainEventTargetType = EventTarget & HTMLInputElement & {
  name: keyof DomainConfig;
};
