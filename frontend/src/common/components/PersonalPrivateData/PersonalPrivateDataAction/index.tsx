import { Button } from '@mui/material';
import { FC } from 'react';

import useStyles from './styles';

interface Props {
  fileLink: string;
}

const PersonalPrivateDataAction: FC<Props> = ({ fileLink }) => {
  const classes = useStyles();
  return (
    <Button className={classes.root} variant="contained" color="primary">
      <a
        className={classes.downloadLink}
        download="vpnhouse.conf"
        href={fileLink}
        id="download-as-file-link"
      >
        Save config to file
      </a>
    </Button>
  );
};

export default PersonalPrivateDataAction;
