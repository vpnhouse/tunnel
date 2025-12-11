// Simplified types for DatePicker

export type PropsType = {
  isEmpty?: boolean;
  label?: string;
  value?: Date | null;
  onChange?: (date: Date | null) => void;
  minDate?: Date;
  maxDate?: Date;
  disabled?: boolean;
  [key: string]: unknown;
}

export type StylesPropsType = {
  isEmpty: boolean;
};
