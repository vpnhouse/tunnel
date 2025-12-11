import { ReactNode } from 'react';
import { TextFieldProps as MuiTextFieldProps } from '@mui/material/TextField';
import { SxProps, Theme } from '@mui/material/styles';

export type TextFieldProps = MuiTextFieldProps & {
  endAdornment?: ReactNode;
  sx?: SxProps<Theme>;
}
