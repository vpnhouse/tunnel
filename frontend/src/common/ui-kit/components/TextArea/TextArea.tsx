import { FC, useCallback, useState } from 'react';
import Typography from '@mui/material/Typography';
import ExpandMore from '@mui/icons-material/ExpandMore';
import ExpandLess from '@mui/icons-material/ExpandLess';
import Box from '@mui/material/Box';

import { CopyToClipboardButton, IconButton } from '../index';
import { PropsType } from './TextArea.types';

const TextArea: FC<PropsType> = ({
  value,
  tableView = false
}) => {
  const [allVisible, setAllVisible] = useState(false);

  const toogleVisibilityHandler = useCallback(() => {
    setAllVisible((prevState) => !prevState);
  }, []);

  return (
    <Box
      sx={{
        display: 'flex',
        justifyContent: 'space-between',
        width: tableView ? '536px' : 'unset'
      }}
    >
      <Typography
        variant="body1"
        sx={{
          whiteSpace: 'pre-line',
          display: '-webkit-box',
          WebkitLineClamp: 3,
          WebkitBoxOrient: allVisible ? 'unset' : 'vertical',
          overflow: 'hidden',
          overflowWrap: 'anywhere',
          height: '100%',
          fontSize: '14px'
        }}
      >
        {value}
      </Typography>
      <Box
        sx={{
          display: 'flex',
          flexDirection: 'column',
          marginLeft: '8px',
          marginRight: '-10px'
        }}
      >
        <IconButton
          color="primary"
          onClick={toogleVisibilityHandler}
          icon={allVisible ? ExpandLess : ExpandMore}
          title={allVisible ? 'Hide' : 'Expand'}
          iconProps={{
            fontSize: '30px'
          }}
        />
        <CopyToClipboardButton value={value} />
      </Box>
    </Box>
  );
};

export default TextArea;
