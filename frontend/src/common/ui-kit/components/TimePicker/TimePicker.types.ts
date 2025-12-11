// Simplified types for TimePicker

export type PropsType = {
  isEmpty?: boolean;
  label?: string;
  value?: Date | null;
  onChange?: (time: Date | null) => void;
  disabled?: boolean;
  [key: string]: unknown;
}

export type StylesPropsType = {
  isEmpty: boolean;
};
