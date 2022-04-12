import React, { FC, useCallback, useState } from 'react';
import { Typography } from '@material-ui/core';
import { ExpandMore, ExpandLess } from '@material-ui/icons';

import { CopyToClipboardButton, IconButton } from '../index';
import { PropsType } from './TextArea.types';
import useStyles from './TextArea.styles';

const TextArea: FC<PropsType> = ({
  value,
  tableView = false
}) => {
  const [allVisible, setAllVisible] = useState(false);
  const classes = useStyles({
    allVisible,
    tableView
  });

  const toogleVisibilityHandler = useCallback(() => {
    setAllVisible((prevState) => !prevState);
  }, []);

  return (
    <div className={classes.root}>
      <Typography variant="body1" className={classes.textArea}>
        {value}
      </Typography>
      <div className={classes.actions}>
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
      </div>
    </div>
  );
};

export default TextArea;
