import { TextFieldProps } from '@mui/material/TextField';

export type PropsType = Partial<TextFieldProps> & {
  fieldName: string;
  delimiter: string;
  compoundValue: string;
  labels: string[];
  fieldWidth?: FieldWithType[];
  onFieldsChange: (field: string, value: string) => void;
}

export type FieldWithType = 'narrow' | 'normal' | 'wide';

export type ValuesType = {
  [index: string]: string;
}

export type StylesPropsType = {
  width: FieldWithType;
}
