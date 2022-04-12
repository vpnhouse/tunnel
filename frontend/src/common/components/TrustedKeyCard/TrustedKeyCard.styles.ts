import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    paper: {
      width: '636px',
      padding: '48px 32px 28px',
      margin: '0 auto 16px',
      position: 'relative'
    },
    paperError: {
      backgroundColor: palette.error.dark
    },
    actions: {
      position: 'absolute',
      left: '610px',
      top: '18px'
    },
    form: {
      '& > div:last-child': {
        marginBottom: 0
      }
    },
    buttonLine: {
      display: 'flex',
      justifyContent: 'flex-end',
      '& > :not(:last-child)': {
        marginRight: '16px'
      }
    },
    commonError: {
      display: 'flex',
      padding: '8px 0 12px',
      alignItems: 'end',
      '& svg': {
        display: 'block',
        height: '15px',
        marginRight: '5px'
      }
    },
    publicKey: {
      marginBottom: '28px',
      '& .MuiOutlinedInput-inputMultiline': {
        height: '100%',
        fontSize: '14px',
        paddingRight: '1px'
      }
    }
  }));

export default useStyles;
