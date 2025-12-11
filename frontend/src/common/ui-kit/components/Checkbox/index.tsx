
import { FC } from 'react';
import Checkbox, { CheckboxProps } from '@mui/material/Checkbox';

import CheckboxIcon from './assets/ChekboxIcon';
import CheckedIcon from './assets/CheckedIcon';

const CustomCheckbox: FC<CheckboxProps> = (props) => {
  const { checked, ...restProps } = props;
  return (
    <Checkbox
      {...restProps}
      checkedIcon={<CheckedIcon />}
      icon={<CheckboxIcon />}
      checked={checked === undefined ? false : checked}
    />
  );
};

export default CustomCheckbox;
