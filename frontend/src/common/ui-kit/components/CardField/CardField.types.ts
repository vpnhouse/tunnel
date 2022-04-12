import { OutlinedTextFieldProps } from '@material-ui/core/TextField/TextField';
import { KeyboardDatePickerProps, KeyboardTimePickerProps } from '@material-ui/pickers';

export type TextFieldType = {
  type: 'TEXT';
  textprops?: Partial<OutlinedTextFieldProps>;
}

export type TextAreaType = {
  type: 'TEXTAREA';
  textprops?: Partial<OutlinedTextFieldProps>;
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
  datePickerProps?: Partial<KeyboardDatePickerProps>;
  timePickerProps?: Partial<KeyboardTimePickerProps>;
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
  validationError: string;
  serverError: string;
  options?: CardFieldOptionsType;
  loadOptions?: LoadFileOptionsType
}

export type StylesPropsTipe = {
  tableView: boolean;
}
