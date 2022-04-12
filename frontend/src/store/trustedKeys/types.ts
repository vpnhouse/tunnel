import { components } from '@schema';

export type TrustedKeyType = components['schemas']['TrustedKey'];
export type TrustedKeyRecordType = components['schemas']['TrustedKeyRecord'];

export type TrustedKeyErrorType = {
  [key in keyof TrustedKeyRecordType]?: string;
} & {
  common?: string;
}

export type TrustedKeyCardType = {
  trustedKeyInfo: TrustedKeyRecordType;
  serverError?: TrustedKeyErrorType;
  isNotSaved?: boolean;
  isEditing: boolean;
}

export type TrustedKeyStoreType = {
  trustedKeys: TrustedKeyCardType[];
  trustedKeyToSave: TrustedKeyRecordType | null;
}

export type SaveKeyInfoType = Omit<TrustedKeyCardType, 'isEditing'> & {
  prevId: string;
}

export type DeleteKeyType = Pick<TrustedKeyRecordType, 'id'> & Pick<TrustedKeyCardType, 'isNotSaved'>

export type KeySetEditingType = {
  id: string;
  isEditing: boolean;
}
