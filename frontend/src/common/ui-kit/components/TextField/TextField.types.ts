import { ReactNode } from 'react';
import { OutlinedTextFieldProps } from '@material-ui/core/TextField/TextField';

export type PropsType = Partial<OutlinedTextFieldProps> & {
  endAdornment?: ReactNode
}
