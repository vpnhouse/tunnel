import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(() =>
  createStyles({
    root: {
      display: 'flex'
    },
    dnsField: {
      width: 160
    },
    deleteButton: {
      marginLeft: 20,
      marginTop: 6,
      padding: '0 20px'
    }
  }));

export default useStyles;
