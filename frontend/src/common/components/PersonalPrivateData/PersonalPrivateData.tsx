import React, { FC } from 'react';
import QRCode from 'qrcode.react';

import logoImage from '@common/assets/logo-high-resolution.png';

import { PersonalPrivateDataProps } from './PersonalPrivateData.types';
import useStyles from './PersonalPrivateData.styles';

const PersonalPrivateData: FC<PersonalPrivateDataProps> = ({ value }) => {
  const classes = useStyles();

  return (
    <div className={classes.personalPrivateData}>
      <div className={classes.qrCodeWrapper}>
        <QRCode
          value={value}
          renderAs="svg"
          size={224}
          imageSettings={{
            src: logoImage,
            height: 70,
            width: 59,
            excavate: true
          }}
        />
      </div>

      <div className={classes.dataWrapper}><pre>{value}</pre></div>
    </div>
  );
};

export default PersonalPrivateData;
