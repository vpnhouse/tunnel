import { OutlinedTextFieldProps } from '@material-ui/core/TextField/TextField';

export type PropsType = Partial<OutlinedTextFieldProps> & {
  fieldName: string;
  delimiter: string;
  compoundValue: string;
  labels: string[];
  fieldWidth?: FieldWithType[];
  onFieldsChange: (field: string, value: string) => void;
}

export type FieldWithType = 'narrow' | 'normal';

export type ValuesType = {
  [index: string]: string;
}

export type StylesPropsType = {
  width: FieldWithType;
}
