import { FC, ElementType } from 'react';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import SvgIcon from '@mui/material/SvgIcon';
import { styled } from '@mui/material/styles';
import { Link as RouterLink } from 'react-router-dom';

import { PropsType } from './RouteMenuItem.types';

interface StyledProps {
  isSelected?: boolean;
}

const StyledLink = styled(RouterLink, {
  shouldForwardProp: (prop) => prop !== 'isSelected'
})<StyledProps>(({ theme, isSelected }) => ({
  display: 'flex',
  alignItems: 'center',
  padding: '12px 24px',
  textDecoration: 'none',
  color: 'inherit',
  borderLeft: isSelected ? `3px solid ${theme.palette.primary.main}` : '3px solid transparent',
  backgroundColor: isSelected ? 'rgba(255, 255, 255, 0.05)' : 'transparent',
  cursor: 'pointer',
  '&:hover': {
    backgroundColor: 'rgba(255, 255, 255, 0.08)'
  },
  [(theme.breakpoints.down as (key: string) => string)('lg')]: {
    padding: '12px',
    justifyContent: 'center'
  }
}));

const StyledListItemIcon = styled(ListItemIcon)(({ theme }) => ({
  minWidth: 40,
  color: theme.palette.text.primary,
  [(theme.breakpoints.down as (key: string) => string)('lg')]: {
    minWidth: 'auto'
  }
}));

const StyledListItemText = styled(ListItemText)(({ theme }) => ({
  margin: 0,
  '& .MuiListItemText-primary': {
    fontSize: 14,
    fontWeight: 500
  },
  '& .MuiListItemText-secondary': {
    fontSize: 12,
    color: theme.palette.text.secondary
  },
  [(theme.breakpoints.down as (key: string) => string)('lg')]: {
    display: 'none'
  }
}));

const RouteMenuItem: FC<PropsType> = ({
  selected,
  icon: IconComponent,
  route,
  pageTitle,
  extraInfo,
  onClick
}) => {
  return (
    <StyledLink
      to={route}
      onClick={onClick}
      isSelected={selected}
    >
      <StyledListItemIcon>
        <SvgIcon component={IconComponent as ElementType} sx={{ fontSize: 24 }} />
      </StyledListItemIcon>
      <StyledListItemText
        primary={pageTitle}
        secondary={extraInfo}
      />
    </StyledLink>
  );
};

export default RouteMenuItem;
