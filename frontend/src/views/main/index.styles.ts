import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    root: {
      height: '100%',
      display: 'flex',
      backgroundColor: palette.background.default,
      overflowX: 'auto'
    },
    content: {
      flex: '1 1 0'
    }
  }));

export default useStyles;
