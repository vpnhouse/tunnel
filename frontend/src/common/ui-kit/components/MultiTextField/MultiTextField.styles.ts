import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    root: {
      display: 'flex',
      '& > :not(:last-child)': {
        marginRight: '12px'
      }
    },
    narrow: {
      width: '24%',
      flex: '1 0 auto',
      minHeight: '70px'
    },
    normal: {
      minHeight: '70px'
    },
    validationError: {
      color: palette.error.main,
      marginTop: '-22px',
      marginBottom: '5px',
      textAlign: 'end',
      paddingRight: '50%'
    }
  }));

export default useStyles;
