import { FC } from 'react';
import Button, { ButtonProps } from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import { styled } from '@mui/material/styles';

interface Props extends ButtonProps {
  isLoading?: boolean;
}

const StyledButton = styled(Button)(({ theme }) => ({
  height: '56px',
  padding: '0 42px',
  color: theme.palette.text.primary,
  position: 'relative',
  boxShadow: 'none',
  borderRadius: 8,
  transition: 'background-color 0 ease',
  '&:hover': {
    boxShadow: 'none',
    '& path': {
      fill: theme.palette.text.primary
    }
  },
  '&.Mui-disabled': {
    backgroundColor: theme.palette.action.disabled,
    color: theme.palette.text.disabled,
    '& path': {
      fill: theme.palette.text.secondary
    }
  },
  '& path': {
    fill: theme.palette.text.primary
  },
  '&.MuiButton-containedPrimary': {
    '&:hover': {
      backgroundColor: theme.palette.primary.light
    },
    '&:focus': {
      backgroundColor: theme.palette.primary.light
    }
  },
  '&.MuiButton-containedSecondary': {
    '&:hover': {
      backgroundColor: theme.palette.secondary.light
    }
  },
  '&.MuiButton-text': {
    ...theme.typography.subtitle1,
    fill: theme.palette.primary.main,
    '&:hover': {
      backgroundColor: theme.palette.background.paper,
      color: theme.palette.primary.light,
      fill: theme.palette.primary.light
    }
  }
}));

const Spinner = styled(CircularProgress)({
  position: 'absolute'
});

const CustomButton: FC<Props> = ({ children, isLoading, disabled, ...rest }) => {
  return (
    <StyledButton
      {...rest}
      disabled={disabled || isLoading}
    >
      {children}
      {isLoading && <Spinner size={24} />}
    </StyledButton>
  );
};

export default CustomButton;
