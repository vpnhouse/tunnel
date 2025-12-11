// Simplified types for DateTimePicker

export type PropsType = {
  dateLabel: string;
  dateName: string;
  timeLabel: string;
  timeName: string;
  value: string;
  validationError: string;
  onChangeHandler: (date: Date | null, time: Date | null) => void;
  datePickerProps?: Record<string, unknown>;
  timePickerProps?: Record<string, unknown>;
  classNames?: {
    root?: string;
    pickers?: string;
  };
}
