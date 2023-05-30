import * as React from 'react';
import { FC } from 'react';
import { CheckboxProps, Checkbox as MuiCheckbox } from '@material-ui/core';

import CheckboxIcon from './assets/ChekboxIcon';
import CheckedIcon from './assets/CheckedIcon';

const Checkbox: FC<CheckboxProps> = (props) => {
  const { checked, ...restProps } = props;
  return (
    <MuiCheckbox {...restProps} checkedIcon={<CheckedIcon />} icon={<CheckboxIcon />} checked={checked === undefined ? false : checked} />
  );
};

export default Checkbox;
