import { FC } from 'react';
import { QRCodeSVG } from 'qrcode.react';

import logoImage from '@common/assets/logo-high-resolution.png';

import { PersonalPrivateDataProps } from './PersonalPrivateData.types';
import useStyles from './PersonalPrivateData.styles';

const PersonalPrivateData: FC<PersonalPrivateDataProps> = ({ value }) => {
  const classes = useStyles();

  return (
    <div className={classes.personalPrivateData}>
      <div className={classes.qrCodeWrapper}>
        <QRCodeSVG
          value={value}
          size={224}
          level="M"
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
