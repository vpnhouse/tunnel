import { FC } from 'react';
import InputAdornment from '@mui/material/InputAdornment';
import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

import { IconButton } from '@root/common/ui-kit/components';

import VisibilityIcon from './VisibilityIcon';
import { PropsType } from './VisibilityAdornment.types';

const IconWrapper = styled(Box, {
  shouldForwardProp: (prop) => prop !== 'disabled'
})<{ disabled?: boolean }>(({ disabled }) => ({
  opacity: disabled ? 0.5 : 1,
  transition: 'opacity 0.2s ease'
}));

const VisibilityAdornment: FC<PropsType> = ({ showPassword, toggleShowPasswordHandler, tabIndex }) => {
  return (
    <InputAdornment position="end">
      <IconWrapper disabled={!showPassword}>
        <IconButton
          onClick={toggleShowPasswordHandler}
          color="primary"
          tabIndex={tabIndex}
          icon={VisibilityIcon}
        />
      </IconWrapper>
    </InputAdornment>
  );
};

export default VisibilityAdornment;
