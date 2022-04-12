import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(() =>
  createStyles({
    root: {
      '& svg': {
        display: 'block'
      }
    }
  }));

export default useStyles;
