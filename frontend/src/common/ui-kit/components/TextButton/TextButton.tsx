import { FC } from 'react';
import Icon from '@mui/material/Icon';
import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

import { PropsType } from './TextButton.types';

const TextButtonRoot = styled(Box)(({ theme }) => ({
  display: 'flex',
  alignItems: 'center',
  height: '48px',
  padding: '0 12px',
  color: theme.palette.text.primary,
  '&:hover': {
    backgroundColor: theme.palette.background.paper,
    color: theme.palette.primary.light,
    cursor: 'pointer'
  }
}));

const Label = styled('span')(({ theme }) => ({
  ...theme.typography.subtitle1,
  paddingLeft: '8px',
  whiteSpace: 'nowrap'
}));

const TextButton: FC<PropsType> = ({ icon, label, onClick }) => {
  return (
    <TextButtonRoot onClick={onClick}>
      <Icon fontSize="small" component={icon} />
      <Label>
        {label}
      </Label>
    </TextButtonRoot>
  );
};

export default TextButton;
