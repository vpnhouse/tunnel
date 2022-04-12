import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(() =>
  createStyles({
    personalPrivateData: {
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      '& > button': {
        margin: '15px 0'
      }
    },
    qrCodeWrapper: {
      padding: '15px',
      background: '#fff'
    },
    dataWrapper: {
      maxWidth: '100%',
      overflow: 'auto',
      '& > pre': {
        margin: 0,
        whiteSpace: 'pre-line'
      }
    }
  }));

export default useStyles;
