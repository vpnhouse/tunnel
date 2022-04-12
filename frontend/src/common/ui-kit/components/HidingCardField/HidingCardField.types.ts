import { OutlinedTextFieldProps } from '@material-ui/core/TextField/TextField';
import { KeyboardDatePickerProps, KeyboardTimePickerProps } from '@material-ui/pickers';

import { FieldWithType } from '../MultiTextField/MultiTextField.types';

export type TextFieldType = {
  type: 'TEXT';
  textprops?: Partial<OutlinedTextFieldProps>;
}

export type TextAreaType = {
  type: 'TEXTAREA';
  textprops?: Partial<OutlinedTextFieldProps>;
}

export type MultiFieldType = {
  type: 'MULTI';
  delimiter: string;
  labels: string[];
  fieldWidth?: FieldWithType[];
  onFieldsChange: (field: string, value: string) => void;
  textprops?: Partial<OutlinedTextFieldProps>;
}

export type DateTimeType = {
  type: 'DATETIME';
  onChangeHandler: (date: Date | null, time: Date | null) => void;
  dateLabel: string;
  dateName: string;
  timeLabel: string;
  timeName: string;
  datePickerProps?: Partial<KeyboardDatePickerProps>;
  timePickerProps?: Partial<KeyboardTimePickerProps>;
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
  validationError: string;
  serverError: string;
  onRemoveFieldHandler: (fieldName: string) => void;
  options?: CardFieldOptionsType
  loadOptions?: LoadFileOptionsType
}
