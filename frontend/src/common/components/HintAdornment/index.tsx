import { InputAdornment, Tooltip } from '@material-ui/core';
import { HelpOutlineRounded } from '@material-ui/icons';
import React, { FC } from 'react';

import useStyles from './styles';

interface Props {
  text: string;
}

const HintAdornment: FC<Props> = ({ text }) => {
  const classes = useStyles();

  return (
    <InputAdornment position="end">
      <Tooltip title={text}>
        <HelpOutlineRounded className={classes.root} />
      </Tooltip>
    </InputAdornment>
  );
};

export default HintAdornment;
