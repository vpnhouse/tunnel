import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    root: {
      height: '100%',
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      backgroundColor: palette.background.default
    },
    enterGroup: {
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center'
    },
    logo: {
      width: 130,
      height: 32,
      marginBottom: 32
    },
    form: {
      width: '320px'
    },
    passwordInput: {
      marginBottom: 12,
      marginTop: 0
    },
    restartButton: {
      width: '221px'
    }
  }));

export default useStyles;
