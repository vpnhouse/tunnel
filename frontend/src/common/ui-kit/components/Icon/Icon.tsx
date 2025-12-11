import { FC } from 'react';
import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

import * as Icons from '@common/ui-kit/icons';

import { PropsType } from './Icon.types';

const IconWrapper = styled(Box)({
  display: 'inline-flex',
  '& svg': {
    width: '24px',
    height: '24px'
  }
});

const Icon: FC<PropsType> = ({ icon, className = '' }) => {
  return (
    <IconWrapper
      className={className}
      dangerouslySetInnerHTML={{ __html: Icons[icon] }}
    />
  );
};

export default Icon;
