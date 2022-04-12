import { GET_IPV4, GET_WIREGUARD, PEERS } from '@constants/apiPaths';
import { FlatPeerType } from '@root/store/peers/types';

const { API_URL } = process.env;

export const PEERS_URL = `${API_URL}${PEERS}`;
export const PEERS_WIREGUARD_URL = `${API_URL}${GET_WIREGUARD}`;
export const GET_IPV4_URL = `${API_URL}${GET_IPV4}`;
export const EMPTY_PEER: FlatPeerType = {
  id: 0
};
