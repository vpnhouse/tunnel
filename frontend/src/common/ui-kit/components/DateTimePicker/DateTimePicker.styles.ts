import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    root: {
      flex: '1 0 auto',
      position: 'relative'
    },
    pickers: {
      display: 'flex',
      justifyContent: 'space-between'
    },
    validationError: {
      color: palette.error.main,
      marginBottom: '5px',
      textAlign: 'end',
      paddingRight: '14px',
      position: 'absolute',
      right: 0
    }
  }));

export default useStyles;
