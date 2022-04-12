import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ typography, palette }) =>
  createStyles({
    section: {
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      background: palette.background.default,
      height: '100vh'
    },
    title: {
      color: '#fff',
      fontFamily: typography.fontFamily,
      margin: '15px 0 0',
      fontSize: '32px'
    }
  }));

export default useStyles;
