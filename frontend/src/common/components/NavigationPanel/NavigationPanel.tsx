import React, { FC } from 'react';
import { Link } from 'react-router-dom';

import logo from '@common/assets/logo.png';
import logoMobile from '@common/assets/logo-mobile.png';

import { Menu } from '../index';
import useStyles from './NavigationPanel.styles';


const NavigationPanel: FC = () => {
  const classes = useStyles();

  return (
    <div className={classes.root}>
      <Link to="/">
        <img className={classes.logo} src={logo} alt="VPNHouse" />
        <img className={classes.logo__mobile} src={logoMobile} alt="VPNHouse" />
      </Link>
      <Menu />
    </div>
  );
};

export default NavigationPanel;
