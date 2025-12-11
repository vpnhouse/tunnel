import { TextFieldProps as MuiTextFieldProps } from '@mui/material/TextField';

import { FieldWithType } from '../MultiTextField/MultiTextField.types';

export type TextFieldType = {
  type: 'TEXT';
  textprops?: Partial<MuiTextFieldProps>;
}

export type TextAreaType = {
  type: 'TEXTAREA';
  textprops?: Partial<MuiTextFieldProps>;
}

export type MultiFieldType = {
  type: 'MULTI';
  delimiter: string;
  labels: string[];
  fieldWidth?: FieldWithType[];
  onFieldsChange: (field: string, value: string) => void;
  textprops?: Partial<MuiTextFieldProps>;
}

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

export type CardFieldOptionsType =
  TextFieldType
  | DateTimeType
  | TextAreaType
  | MultiFieldType;

export type LoadFileOptionsType = {
  accept: string;
  onLoad: (name: string, value: string) => void;
}

export type PropsType = {
  isEditing: boolean;
  readonly?: boolean;
  header?: boolean;
  copyToClipboard?: boolean;
  name: string;
  label: string;
  value: string;
  validationError?: string;
  serverError?: string;
  onRemoveFieldHandler: (fieldName: string) => void;
  options?: CardFieldOptionsType
  loadOptions?: LoadFileOptionsType
}
