import { ElementType, MouseEvent } from 'react';

export type PropsType = {
  selected: boolean;
  icon: ElementType;
  route: string;
  pageTitle: string;
  extraInfo?: string
  onClick?: (e: MouseEvent) => void;
}

export type SelectedProps = {
  selected: boolean;
}
