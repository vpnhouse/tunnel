import React, { FC, useCallback, useEffect, useRef } from 'react';
import { useStore } from 'effector-react';
import { Typography } from '@material-ui/core';
import { AddCircle, Autorenew } from '@material-ui/icons';

import { TrustedKeyCard } from '@common/components';
import { Button } from '@common/ui-kit/components';
import {
  $trustedKeysStore, cancelCreateTrustedKey,
  createTrustedKey,
  getAllTrustedKeysFx
} from '@root/store/trustedKeys';

import useStyles from './index.styles';

const TrustedKeys: FC = () => {
  const classes = useStyles();
  const { trustedKeys, trustedKeyToSave } = useStore($trustedKeysStore);
  const topPageRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    getAllTrustedKeysFx();
  }, []);

  const refreshPageHandler = useCallback(() => {
    getAllTrustedKeysFx();
    cancelCreateTrustedKey();
  }, []);

  const addKeyHandler = useCallback(() => {
    !trustedKeyToSave && createTrustedKey();
    topPageRef?.current?.scrollIntoView();
  }, [trustedKeyToSave]);

  return (
    <div className={classes.root}>
      <div className={classes.header}>
        <Typography variant="h1" color="textPrimary">
          Trusted Keys
        </Typography>
        <div className={classes.actions}>
          <Button
            variant="contained"
            color="secondary"
            startIcon={<Autorenew />}
            onClick={refreshPageHandler}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            color="primary"
            startIcon={<AddCircle />}
            onClick={addKeyHandler}
          >
            Add new
          </Button>
        </div>
      </div>
      <div className={classes.main}>
        <div ref={topPageRef} />
        {!!trustedKeyToSave && (
          <TrustedKeyCard
            key="newKey"
            isEditing
            isNotSaved
            trustedKeyInfo={trustedKeyToSave}
          />
        )}
        {trustedKeys.map((item, index) => (
          <TrustedKeyCard
            /* In case if user tries save keys with existing id, cards mist have different keys */
            key={item.isNotSaved ? `notSaved-${index}-${item.trustedKeyInfo.id}}` : item.trustedKeyInfo.id}
            {...item}
          />
        ))}
      </div>
    </div>
  );
};

export default TrustedKeys;
