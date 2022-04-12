import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(() =>
  createStyles({
    root: {
      width: '100%',
      flex: '1 0 auto',
      paddingTop: 0,
      paddingBottom: 0,
      position: 'relative'
    }
  }));

export default useStyles;
