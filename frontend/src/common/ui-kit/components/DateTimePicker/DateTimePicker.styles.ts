import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    root: {
      flex: '1 0 auto'
    },
    pickers: {
      display: 'flex',
      justifyContent: 'space-between'
    },
    validationError: {
      color: palette.error.main,
      marginTop: '-22px',
      marginBottom: '5px',
      textAlign: 'end',
      paddingRight: '14px'
    }
  }));

export default useStyles;
