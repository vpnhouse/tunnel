import { TrustedKeyErrorType, TrustedKeyRecordType } from '@root/store/trustedKeys/types';

export type PropsType = {
  trustedKeyInfo: TrustedKeyRecordType;
  serverError?: TrustedKeyErrorType;
  isEditing: boolean;
  isNotSaved?: boolean;
};

export type TrustedKeysFieldsType = keyof TrustedKeyRecordType;
export type TrustedKeysEventTargetType = EventTarget & HTMLInputElement & {
  name: TrustedKeysFieldsType;
};

export type TrustedKeysPatternsType = {
  [key in TrustedKeysFieldsType]?: RegExp;
};

export type TrustedKeysValidationType = {
  [key in TrustedKeysFieldsType]: (field: TrustedKeysFieldsType, value: string) => string;
};

export type PatternErrorType = {
  [key in TrustedKeysFieldsType]?: string;
} & {
  required: string;
};
