import { ReactElement } from 'react';

export type DialogType = {
  title: string;
  message: string | ReactElement;
  successButtonTitle?: string;
  onlyClose?: boolean;
  successButtonHandler?: () => void;
}
