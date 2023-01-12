import { KeyboardTimePickerProps, KeyboardDatePickerProps } from '@material-ui/pickers';

export type PropsType = {
  dateLabel: string;
  dateName: string;
  timeLabel: string;
  timeName: string;
  value: string;
  validationError: string;
  onChangeHandler: (date: Date | null, time: Date | null) => void;
  datePickerProps?: Partial<KeyboardDatePickerProps>;
  timePickerProps?: Partial<KeyboardTimePickerProps>;
  classNames?: {
    root?: string;
    pickers?: string;
  };
}
