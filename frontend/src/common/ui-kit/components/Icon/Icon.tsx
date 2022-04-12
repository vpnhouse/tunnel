import React, { FC } from 'react';

import * as Icons from '@common/ui-kit/icons';

import { PropsType } from './Icon.types';
import useStyles from './Icon.styles';

const Icon: FC<PropsType> = ({ icon, className = '' }) => {
  const classes = useStyles();

  return (
    <span className={`${classes.root} ${className}`} dangerouslySetInnerHTML={{ __html: Icons[icon] }} />
  );
};

export default Icon;
