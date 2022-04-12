import React, { FC, useCallback, useState } from 'react';
import QRCode from 'qrcode.react';

import { Button } from '@common/ui-kit/components';

import { PersonalPrivateDataProps } from './PersonalPrivateData.types';
import useStyles from './PersonalPrivateData.styles';

const PersonalPrivateData: FC<PersonalPrivateDataProps> = ({ value }) => {
  const classes = useStyles();
  const [isVisible, setIsVisible] = useState(false);

  const toggleData = useCallback(() => {
    setIsVisible((prevState) => !prevState);
  }, []);

  return (
    <div className={classes.personalPrivateData}>
      <div className={classes.qrCodeWrapper}>
        <QRCode value={value} renderAs="svg" size={256} />
      </div>
      <Button onClick={toggleData} variant="contained" color="primary">
        {isVisible ? 'Hide config for desktop clients' : 'Show config for desktop clients'}
      </Button>
      {isVisible && <div className={classes.dataWrapper}><pre>{value}</pre></div>}
    </div>
  );
};

export default PersonalPrivateData;
