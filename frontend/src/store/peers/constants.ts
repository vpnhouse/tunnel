import { GET_IPV4, GET_WIREGUARD, PEERS } from '@constants/apiPaths';
import { FlatPeerType } from '@root/store/peers/types';

export const PEERS_URL = PEERS;
export const PEERS_WIREGUARD_URL = GET_WIREGUARD;
export const GET_IPV4_URL = GET_IPV4;
export const EMPTY_PEER: FlatPeerType = {
  id: 0
};
