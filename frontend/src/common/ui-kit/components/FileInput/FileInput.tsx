import React, { FC, ChangeEvent, useCallback } from 'react';
import { InsertDriveFile } from '@material-ui/icons';

import { IconButton } from '@common/ui-kit/components';

import { PropsType } from './FileInput.types';
import useStyles from './FileInput.styles';

const FileInput: FC<PropsType> = ({ accept, name, onLoad }) => {
  const classes = useStyles();

  const fileInputRef = React.useRef<HTMLInputElement>(null);

  const onLoadFileHandler = useCallback(() => {
    fileInputRef.current?.click();
  }, []);

  const onReadFileHandler = useCallback(async (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const file = e.target.files[0];
      const fileText = await file.text();
      onLoad?.(name, fileText);
    }
  }, [name, onLoad]);

  return (
    <>
      <input
        accept={accept}
        className={classes.input}
        id="load-file"
        type="file"
        ref={fileInputRef}
        onChange={onReadFileHandler}
      />
      <label htmlFor="load-file">
        <IconButton
          color="primary"
          icon={InsertDriveFile}
          onClick={onLoadFileHandler}
          title="Load from file"
          iconProps={{
            fontSize: '22px'
          }}
        />
      </label>
    </>

  );
};

export default FileInput;
