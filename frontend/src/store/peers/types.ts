import { components } from '@schema';

type PeerIdentifiersType = components['schemas']['ConnectionIdentifiers'];
export type PeerType = components['schemas']['Peer'];
export type PeerRecordType = components['schemas']['PeerRecord'];

export type FlatPeerType = Pick<PeerRecordType, 'id'>
  & Omit<PeerType, 'info_wireguard' | 'identifiers'>
  & PeerIdentifiersType
  & PeerKeys
  & PeersWireguard;

export type PeerErrorType = {
  [key in keyof Partial<FlatPeerType>]: string;
} & {
  common?: string;
}

export type PeersWireguard = {
  peerData: {
    private_key: string;
    ipv4: string;
    label?: string | null;
  };
  dns: string[];
  server_public_key: string;
  allowed_ips: string[];
  server_ipv4: string;
  server_port: string;
  keepalive: number;
}

export type PeerKeys = {
  public_key: string;
  private_key: string;
}

export type PeerInfoType = {
  peerInfo: FlatPeerType;
  serverError?: PeerErrorType;
  isEditing: boolean;
}

export type PeerStoreType = {
  peers: PeerInfoType[];
  peerToSave: FlatPeerType | null;
}

export type PeerSetEditingType = {
  id: number;
  isEditing: boolean;
}
