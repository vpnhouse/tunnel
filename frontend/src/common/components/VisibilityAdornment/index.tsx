import React, { FC } from 'react';
import { InputAdornment } from '@material-ui/core';
import clsx from 'clsx';

import { IconButton } from '@root/common/ui-kit/components';

import VisibilityIcon from './VisibilityIcon';
import { PropsType } from './VisibilityAdornment.types';
import useStyles from './styles';

const VisibilityAdornment: FC<PropsType> = ({ showPassword, toggleShowPasswordHandler, tabIndex }) => {
  const classes = useStyles();

  return (
    <InputAdornment position="end">
      <IconButton
        onClick={toggleShowPasswordHandler}
        color="primary"
        tabIndex={tabIndex}
        className={clsx(classes.root, !showPassword && classes.disabled)}
        icon={VisibilityIcon}
      />
    </InputAdornment>
  );
};

export default VisibilityAdornment;
