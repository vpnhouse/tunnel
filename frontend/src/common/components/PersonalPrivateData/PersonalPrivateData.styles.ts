import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(() =>
  createStyles({
    personalPrivateData: {
      display: 'flex',
      alignItems: 'center'
    },
    qrCodeWrapper: {
      padding: 16,
      background: '#fff',
      borderRadius: 8,
      marginRight: 6,
      maxHeight: 224
    },
    dataWrapper: {
      maxWidth: '100%',
      overflow: 'auto',
      lineHeight: '24px',
      marginTop: -24,
      paddingRight: 24,
      '& > pre': {
        margin: 0
      }
    }
  }));

export default useStyles;
