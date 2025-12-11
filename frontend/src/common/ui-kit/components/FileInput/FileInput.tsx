import { FC, ChangeEvent, useCallback, useRef } from 'react';
import InsertDriveFile from '@mui/icons-material/InsertDriveFile';
import { styled } from '@mui/material/styles';

import { IconButton } from '@common/ui-kit/components';

import { PropsType } from './FileInput.types';

const HiddenInput = styled('input')({
  display: 'none'
});

const FileInput: FC<PropsType> = ({ accept, name, onLoad }) => {
  const fileInputRef = useRef<HTMLInputElement>(null);

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
      <HiddenInput
        accept={accept}
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
