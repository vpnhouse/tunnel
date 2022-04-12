import { ElementType, MouseEvent } from 'react';

export type iconsClassType = 'primary' | 'error';
export type IconPropsType = {
  fontSize?: string;
}

export type PropsType = {
  color: iconsClassType;
  onClick: (event: MouseEvent<HTMLButtonElement>) => void;
  icon: ElementType;
  title?: string;
  className?: string;
  iconProps?: IconPropsType;
  tabIndex?: string;
}
