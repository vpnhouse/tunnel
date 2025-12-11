import { FC } from 'react';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';
import { styled } from '@mui/material/styles';
import { TextFieldProps } from './TextField.types';

const StyledTextField = styled(TextField)(({ theme }) => ({
  display: 'block',
  minHeight: 56,
  '& .MuiInputLabel-root': {
    color: theme.palette.text.disabled,
    top: 2,
    left: 4,
    '&.Mui-error': {
      color: theme.palette.text.disabled
    }
  },
  '& .MuiOutlinedInput-root': {
    borderRadius: 8,
    backgroundColor: '#2B3142',
    '&:hover': {
      backgroundColor: '#3B3F63',
      '&.Mui-disabled': {
        backgroundColor: theme.palette.action.disabled
      }
    },
    '&.Mui-disabled': {
      backgroundColor: theme.palette.action.disabled
    },
    '&.Mui-focused': {
      backgroundColor: '#3B3F63'
    },
    '&.Mui-error': {
      backgroundColor: theme.palette.error.dark
    },
    '& .MuiOutlinedInput-notchedOutline': {
      borderColor: 'transparent'
    },
    '&:hover .MuiOutlinedInput-notchedOutline': {
      borderColor: 'transparent'
    },
    '&.Mui-focused .MuiOutlinedInput-notchedOutline': {
      borderColor: theme.palette.primary.main
    }
  },
  '& .MuiFilledInput-root': {
    borderRadius: 8,
    paddingLeft: 4,
    caretColor: theme.palette.primary.main,
    backgroundColor: '#2B3142',
    '&:hover': {
      backgroundColor: '#3B3F63',
      '&.Mui-disabled': {
        backgroundColor: theme.palette.action.disabled
      }
    },
    '&.Mui-disabled': {
      backgroundColor: theme.palette.action.disabled
    },
    '&.Mui-focused': {
      backgroundColor: '#3B3F63'
    },
    '&.Mui-error': {
      backgroundColor: theme.palette.error.dark,
      caretColor: 'unset'
    },
    '&::before, &::after': {
      display: 'none'
    }
  },
  '& .MuiInputBase-input': {
    color: theme.palette.text.primary,
    '&:-webkit-autofill': {
      WebkitTextFillColor: theme.palette.text.primary,
      transition: 'background-color 5000s ease-in-out 0s'
    },
    '&.Mui-disabled': {
      color: theme.palette.text.disabled,
      WebkitTextFillColor: theme.palette.text.disabled
    }
  },
  '& .MuiFormHelperText-root': {
    textAlign: 'end'
  }
}));

const CustomTextField: FC<TextFieldProps> = ({ endAdornment, variant = 'outlined', ...props }) => {
  return (
    <StyledTextField
      {...props}
      variant={variant}
      margin="dense"
      slotProps={{
        input: {
          endAdornment: endAdornment ? (
            <InputAdornment position="end">{endAdornment}</InputAdornment>
          ) : undefined
        }
      }}
    />
  );
};

export default CustomTextField;
