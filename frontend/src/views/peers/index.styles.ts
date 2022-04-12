import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    root: {
      height: '100%',
      display: 'flex',
      flexDirection: 'column'
    },
    header: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      width: '100%',
      maxWidth: 1072,
      margin: '36px 0',
      '-webkit-user-select': 'none',
      '-moz-user-select': 'none',
      '-ms-user-select': 'none',
      userSelect: 'none'
    },
    refreshButton: {
      width: 160,
      '& svg': {
        height: 16,
        width: 16
      }
    },
    addButton: {
      width: 160,
      '& svg': {
        height: 14,
        width: 14
      }
    },
    main: {
      height: '100%',
      overflow: 'auto',
      display: 'flex',
      flexDirection: 'column',
      width: '100%'
    },
    actions: {
      display: 'flex',
      justifyContent: 'flex-end',
      '& button:not(:last-child)': {
        marginRight: '12px'
      },
      '& button': {
        padding: '0 28px'
      }
    }
  }));

export default useStyles;
