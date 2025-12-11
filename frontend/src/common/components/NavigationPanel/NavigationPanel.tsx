import { FC } from 'react';
import { Link } from 'react-router-dom';
import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

import logo from '@common/assets/logo.png';
import logoMobile from '@common/assets/logo-mobile.png';

import { Menu } from '../index';

const NavRoot = styled(Box)(({ theme }) => ({
  display: 'flex',
  flexDirection: 'column',
  height: '100%',
  width: 260,
  padding: '32px 0',
  backgroundColor: theme.palette.background.paper,
  borderRadius: '0 12px 12px 0',
  [theme.breakpoints.down('lg')]: {
    width: 80
  }
}));

const Logo = styled('img')(({ theme }) => ({
  display: 'block',
  width: 180,
  height: 'auto',
  margin: '0 auto 24px',
  [theme.breakpoints.down('lg')]: {
    display: 'none'
  }
}));

const LogoMobile = styled('img')(({ theme }) => ({
  display: 'none',
  width: 40,
  height: 'auto',
  margin: '0 auto 24px',
  [theme.breakpoints.down('lg')]: {
    display: 'block'
  }
}));

const NavigationPanel: FC = () => {
  return (
    <NavRoot>
      <Link to="/">
        <Logo src={logo} alt="VPNHouse" />
        <LogoMobile src={logoMobile} alt="VPNHouse" />
      </Link>
      <Menu />
    </NavRoot>
  );
};

export default NavigationPanel;
