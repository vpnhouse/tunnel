import { ReactNode } from 'react';
import { OutlinedTextFieldProps } from '@material-ui/core/TextField/TextField';

export type TextFieldProps = Partial<OutlinedTextFieldProps> & {
  endAdornment?: ReactNode
}
