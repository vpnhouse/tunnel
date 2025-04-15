import { FlatPeerType, PeerErrorType } from '@root/store/peers/types';

export type PropsType = {
  peerInfo: FlatPeerType,
  serverError?: PeerErrorType;
  isModal?: boolean;
  open?: boolean;
  onClose?: () => void;
}

export type PeerCardFieldsType = keyof FlatPeerType;
export type PeerCardEventTargetType = EventTarget & HTMLInputElement & {
  name: PeerCardFieldsType;
};

export type PeerCardPatternsType = {
  [key: string]: RegExp;
};

export type PeerCardsValidationType = {
  [key: string]: (field: string, value: string) => string;
};

export type PatternErrorType = {
  [key: string]: string;
} & {
  required: string;
};
