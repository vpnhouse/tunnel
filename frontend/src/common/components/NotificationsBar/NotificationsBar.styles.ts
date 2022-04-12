import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ zIndex, breakpoints }) =>
  createStyles({
    stack: {
      position: 'absolute',
      bottom: '27px',
      zIndex: zIndex.snackbar,
      display: 'flex',
      flexDirection: 'column-reverse',
      alignItems: 'center',
      width: '100%'
    },
    authShift: {
      [breakpoints.up('md')]: {
        left: '456px',
        width: 'calc(100% - 456px)'
      },
      [breakpoints.only('md')]: {
        left: '384px',
        width: 'calc(100% - 384px)'
      }
    }
  }));

export default useStyles;
