import { FC } from 'react';
import InputAdornment from '@mui/material/InputAdornment';
import Tooltip from '@mui/material/Tooltip';
import HelpOutlineRounded from '@mui/icons-material/HelpOutlineRounded';
import { styled } from '@mui/material/styles';

interface Props {
  text: string;
}

const StyledHelpIcon = styled(HelpOutlineRounded)(({ theme }) => ({
  color: theme.palette.text.secondary,
  cursor: 'pointer',
  '&:hover': {
    color: theme.palette.primary.main
  }
}));

const HintAdornment: FC<Props> = ({ text }) => {
  return (
    <InputAdornment position="end">
      <Tooltip title={text}>
        <StyledHelpIcon />
      </Tooltip>
    </InputAdornment>
  );
};

export default HintAdornment;
