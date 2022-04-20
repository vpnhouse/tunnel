import { FlatPeerType, PeerErrorType } from '@root/store/peers/types';

import { patternRequiredValidation, patternValidation } from './PeerCard.utils';
import {
  PatternErrorType,
  PeerCardPatternsType,
  PeerCardsValidationType
} from './PeerCard.types';

export const PEER_FIELD_CAN_BE_NULL: Array<keyof Partial<FlatPeerType>> = [
  'label',
  'claims',
  'ipv4',
  'expires',
  'user_id',
  'installation_id',
  'session_id'
];

export const INVALID_SYMBOLS: PeerCardPatternsType = {
  public_key: /[^A-Za-z0-9+=/]/,
  ipv4: /[^0-9.]/,
  session_id: /[^0-9a-fA-F-]/,
  installation_id: /[^0-9a-fA-F-]/
};

export const SYMBOL_ERRORS: PeerErrorType = {
  public_key: 'Only letters, digits and symbols +/= are allowed',
  ipv4: 'Only digits and dots are allowed',
  session_id: 'Only hexadecimal digits and symbol - are allowed',
  installation_id: 'Only hexadecimal digits and symbol - are allowed'
};

export const PATTERNS: PeerCardPatternsType = {
  public_key: /^[A-Za-z0-9+/]+={0,2}$/,
  ipv4: /^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/,
  session_id: /^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$/,
  installation_id: /^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$/
};

export const PATTERN_ERRORS: PatternErrorType = {
  required: 'This field is required',
  public_key: 'Invalid public key format',
  ipv4: 'Invalid IPV4 format',
  session_id: 'Invalid UUID format',
  installation_id: 'Invalid UUID format'
};

export const PATTERN_VALIDATION: PeerCardsValidationType = {
  public_key: patternRequiredValidation,
  ipv4: patternValidation
};

export const FAQ_CREATE_PEER_IPV4 = 'The address has been randomly picked up from the configured range, you can use it or change it if it is needed';
