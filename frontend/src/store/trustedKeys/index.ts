import { createStore, createEffect, createEvent } from 'effector';

import { fetchData } from '../utils';
import {
  DeleteKeyType,
  KeySetEditingType,
  SaveKeyInfoType,
  TrustedKeyCardType,
  TrustedKeyRecordType,
  TrustedKeyStoreType,
  TrustedKeyType
} from './types';
import { TRUSTED_URL } from './constants';

const initialTrustedKeysStore: TrustedKeyStoreType = {
  trustedKeys: [],
  trustedKeyToSave: null
};

export const $trustedKeysStore = createStore(initialTrustedKeysStore);

export const setTrustedKeys = createEvent<TrustedKeyRecordType[]>();
export const createTrustedKey = createEvent();
export const cancelCreateTrustedKey = createEvent();
export const saveTrustedKey = createEvent<TrustedKeyCardType>();
export const changeTrustedKey = createEvent<TrustedKeyCardType>();
export const deleteTrustedKey = createEvent<DeleteKeyType>();
export const setIsEditing = createEvent<KeySetEditingType>();

export const getAllTrustedKeysFx = createEffect<void, TrustedKeyRecordType[], Response>(
  () => fetchData(TRUSTED_URL).then((res) => res.json())
);

export const saveTrustedKeyFx = createEffect<SaveKeyInfoType, TrustedKeyType, Response>(
  (newKey) => {
    const { id, key } = newKey.trustedKeyInfo;

    return fetchData(
      `${TRUSTED_URL}/${id}`,
      {
        method: 'POST',
        body: key
      }
    )
      .then((res) => res.text());
  }
);

export const changeTrustedKeyFx = createEffect<TrustedKeyRecordType, TrustedKeyType, Response>(
  ({ id, key }) => fetchData(
    `${TRUSTED_URL}/${id}`,
    {
      method: 'PUT',
      body: key
    }
  )
    .then((res) => res.text())
);

export const deleteTrustedKeyFx = createEffect<DeleteKeyType, Response | string, Response>(
  ({ isNotSaved, id }) => {
    if (isNotSaved) return 'Key deleted';

    return fetchData(
      `${TRUSTED_URL}/${id}`,
      {
        method: 'DELETE'
      }
    );
  }
);
