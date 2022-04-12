import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(() =>
  createStyles({
    dnsBlock: {
      marginBottom: '28px'
    },
    addNewField: {
      width: 160,
      backgroundColor: '#2B3142',
      borderRadius: 8,
      '&:hover': {
        backgroundColor: '#3B3F63'
      },
      '& *': {
        cursor: 'text'
      }
    }
  }));

export default useStyles;
