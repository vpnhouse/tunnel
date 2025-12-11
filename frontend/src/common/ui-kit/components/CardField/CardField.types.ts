import { TextFieldProps as MuiTextFieldProps } from '@mui/material/TextField';

import { TextFieldProps } from '@common/ui-kit/components/TextField/TextField.types';

export type TextFieldType = {
  type: 'TEXT';
  textprops?: Partial<MuiTextFieldProps> & Pick<TextFieldProps, 'endAdornment'>;
}

export type TextAreaType = {
  type: 'TEXTAREA';
  textprops?: Partial<MuiTextFieldProps>;
}

export type CardFieldOptionsType =
  TextFieldType
  | DateTimeType
  | TextAreaType

export type DateTimeType = {
  type: 'DATETIME';
  onChangeHandler: (date: Date | null, time: Date | null) => void;
  dateLabel: string;
  dateName: string;
  timeLabel: string;
  timeName: string;
  datePickerProps?: Record<string, unknown>;
  timePickerProps?: Record<string, unknown>;
}

export type LoadFileOptionsType = {
  accept: string;
  onLoad: (name: string, value: string) => void;
}

export type PropsType = {
  isEditing: boolean;
  readonly?: boolean;
  copyToClipboard?: boolean;
  tableView?: boolean;
  label: string;
  name: string;
  value: string;
  validationError?: string;
  serverError?: string;
  options?: CardFieldOptionsType;
  loadOptions?: LoadFileOptionsType;
  isDisable?: boolean;
  disableControl?: boolean;
}

export type StylesPropsType = {
  tableView: boolean;
}
