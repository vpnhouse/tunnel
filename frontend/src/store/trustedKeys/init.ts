import {
  $trustedKeysStore,
  cancelCreateTrustedKey,
  createTrustedKey,
  changeTrustedKey,
  deleteTrustedKey,
  setTrustedKeys,
  saveTrustedKey,
  saveTrustedKeyFx,
  deleteTrustedKeyFx,
  getAllTrustedKeysFx,
  changeTrustedKeyFx,
  setIsEditing
} from './index';
import { EMPTY_TRUSTED_KEY } from './constants';
import { TrustedKeyErrorType } from './types';
import { addNotification, showServerErrorFx } from '../notifications';

$trustedKeysStore
  .on(setTrustedKeys, (state, savedKeys) => ({
    ...state,
    trustedKeys: savedKeys.map((key) => ({
      trustedKeyInfo: key,
      isEditing: false
    }))
  }))
  .on(createTrustedKey, (state) => ({
    ...state,
    trustedKeyToSave: EMPTY_TRUSTED_KEY
  }))
  .on(cancelCreateTrustedKey, (state) => ({
    ...state,
    trustedKeyToSave: null
  }))
  .on(saveTrustedKey, (store, newKey) => ({
    ...store,
    trustedKeys: [newKey, ...store.trustedKeys],
    trustedKeyToSave: null
  }))
  .on(changeTrustedKey, (store, changedKey) => ({
    ...store,
    trustedKeys: store.trustedKeys.map((item) => (
      item.trustedKeyInfo.id === changedKey.trustedKeyInfo.id ? changedKey : item
    ))
  }))
  .on(deleteTrustedKey, (store, { id, isNotSaved = false }) => ({
    ...store,
    trustedKeys: store.trustedKeys.filter((item) =>
      /** If we want delete unsaved keys only,  */
      !(item.trustedKeyInfo.id === id && (!isNotSaved || item.isNotSaved)))
  }))
  .on(setIsEditing, (store, { id, isEditing }) => ({
    ...store,
    trustedKeys: store.trustedKeys.map((item) => (
      item.trustedKeyInfo.id === id
        ? {
          ...item,
          isEditing
        }
        : item
    ))
  }));

getAllTrustedKeysFx.doneData.watch((result) => {
  setTrustedKeys(result);
});

getAllTrustedKeysFx.failData.watch((error) => {
  showServerErrorFx(error);
});

saveTrustedKeyFx.watch((params) => {
  const { isNotSaved, prevId } = params;

  /**
   * If key was not saved, delete it from store (by old id in case if it was changed),
   * it will be saved again later
   */
  isNotSaved && deleteTrustedKey({
    id: prevId,
    isNotSaved
  });
});

saveTrustedKeyFx.done.watch(({ params }) =>
  saveTrustedKey({
    trustedKeyInfo: params.trustedKeyInfo,
    isNotSaved: false,
    isEditing: false
  }));

saveTrustedKeyFx.fail.watch(({ params, error }) => {
  error.json().then((errorDetails) => {
    const field: keyof TrustedKeyErrorType = errorDetails.field || 'common';
    const serverError: TrustedKeyErrorType = { [field]: `${errorDetails.error} ${errorDetails.details || ''}` };

    /** Save with notSaved flag */
    saveTrustedKey({
      trustedKeyInfo: params.trustedKeyInfo,
      isNotSaved: true,
      isEditing: false,
      serverError
    });
  });
});

changeTrustedKeyFx.done.watch(({ params }) => {
  changeTrustedKey({
    trustedKeyInfo: params,
    isEditing: false
  });
});

changeTrustedKeyFx.fail.watch(({ params, error }) => {
  error.json().then((errorDetails) => {
    const field = errorDetails.field || 'common';
    const serverError = { [field]: `${errorDetails.error} ${errorDetails.details || ''}` };
    changeTrustedKey({
      trustedKeyInfo: params,
      isEditing: false,
      serverError
    });
  });
});

deleteTrustedKeyFx.done.watch(({ params }) => {
  const { id, isNotSaved } = params;
  deleteTrustedKey({
    id,
    isNotSaved
  });
  addNotification({
    type: 'info',
    prefix: 'trustedKeyDeleteInfo',
    message: `Trusted key with UUID ${id} was removed`
  });
});

deleteTrustedKeyFx.failData.watch((error) => {
  showServerErrorFx(error);
});
