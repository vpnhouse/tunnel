import { FC, useCallback } from 'react';
import { FileCopyOutlined } from '@mui/icons-material';

import { addNotification } from '@root/store/notifications';
import { IconButton } from '@common/ui-kit/components';

import { PropsType } from './CopyToClipboardButton.types';

const CopyToClipboardButton: FC<PropsType> = ({ value }) => {
  const copyToClipboardHandler = useCallback(() => {
    addNotification({
      type: 'info',
      prefix: 'copyToClipboard',
      message: 'The field content has been copied to clipboard'
    });
    navigator.clipboard.writeText(value);
  }, [value]);

  return (
    <IconButton
      color="primary"
      onClick={copyToClipboardHandler}
      icon={FileCopyOutlined}
      title="Copy to clipboard"
      iconProps={{
        fontSize: '22px'
      }}
    />
  );
};

export default CopyToClipboardButton;
