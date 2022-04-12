import React, { FC } from 'react';

import logo from '@common/assets/logo.png';

import { Menu } from '../index';
import useStyles from './NavigationPanel.styles';

const NavigationPanel: FC = () => {
  const classes = useStyles();

  return (
    <div className={classes.root}>
      <img className={classes.logo} src={logo} alt="logo" />
      <Menu />
    </div>
  );
};

export default NavigationPanel;
