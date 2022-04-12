import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette, typography }) =>
  createStyles({
    root: {
      marginBottom: '10px',
      padding: '11px 32px',
      color: palette.common.white,
      ...typography.subtitle1,
      width: '760px'
    },
    icon: {
      display: 'none'
    },
    message: {
      padding: 0,
      display: 'flex',
      alignItems: 'center'
    },
    filledError: {
      backgroundColor: palette.error.main
    },
    filledInfo: {
      backgroundColor: palette.info.main
    },
    filledWarning: {
      backgroundColor: palette.info.main
    },
    snackRoot: {
      display: 'none'
    }
  }));

export default useStyles;
