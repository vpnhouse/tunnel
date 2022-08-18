import { ReactElement } from 'react';

export type DialogType = {
  title: string;
  message: string | ReactElement;
  successButtonTitle?: string;
  actionComponent?: ReactElement;
  successButtonHandler?: () => void;
}

export type DialogStore = {
  opened: boolean;
} & DialogType;
