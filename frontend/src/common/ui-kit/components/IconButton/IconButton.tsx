import { FC } from 'react';
import IconButton from '@mui/material/IconButton';
import Icon from '@mui/material/Icon';
import Tooltip from '@mui/material/Tooltip';
import Box from '@mui/material/Box';
import { styled, useTheme } from '@mui/material/styles';
import clsx from 'clsx';

import { PropsType } from './IconButton.types';
import { DEFAULT_ICON_PROPS } from './IconButton.constants';

const StyledIconButton = styled(IconButton)(({ theme }) => ({
  width: 36,
  height: 36,
  '&:hover': {
    cursor: 'pointer',
    '& div': {
      '&:before': {
        visibility: 'visible'
      }
    }
  },
  '&:active': {
    '& div': {
      '&:before': {
        backgroundColor: 'rgba(0, 0, 0, 0.6)'
      }
    }
  }
}));

const IconWrapper = styled(Box)({
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  position: 'relative',
  zIndex: 1,
  '&:before': {
    zIndex: -1,
    content: '" "',
    display: 'block',
    visibility: 'hidden',
    width: 36,
    height: 36,
    borderRadius: '50%',
    position: 'absolute',
    backgroundColor: '#2B3142'
  }
});

const StyledTooltip = styled(Tooltip)(({ theme }) => ({
  '& .MuiTooltip-tooltip': {
    backgroundColor: theme.palette.secondary.main,
    fontFamily: "'Roboto', 'Helvetica', 'Arial', sans-serif",
    fontSize: '12px',
    fontWeight: 400
  }
}));

const CustomIconButton: FC<PropsType> = ({
  color,
  onClick,
  icon,
  title = '',
  iconProps = DEFAULT_ICON_PROPS,
  className,
  tabIndex
}) => {
  const theme = useTheme();

  const colorStyles = {
    primary: {
      color: theme.palette.common.white,
      padding: 0,
      '&:hover': {
        backgroundColor: 'transparent',
        color: theme.palette.primary.main
      }
    },
    error: {
      color: theme.palette.error.main,
      padding: '5px',
      height: '34px',
      '&:hover': {
        backgroundColor: 'transparent',
        color: theme.palette.error.light
      }
    }
  };

  return (
    <Tooltip
      title={title}
      placement="right"
    >
      <StyledIconButton
        sx={colorStyles[color]}
        className={clsx(className)}
        onClick={onClick}
        tabIndex={tabIndex ? parseInt(tabIndex) : undefined}
      >
        <IconWrapper id="iconButton">
          <Icon
            component={icon}
            sx={{
              fontSize: iconProps?.fontSize || '24px',
              width: 12,
              height: 12
            }}
          />
        </IconWrapper>
      </StyledIconButton>
    </Tooltip>
  );
};

export default CustomIconButton;
