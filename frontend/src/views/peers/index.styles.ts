import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    root: {
      height: '100%',
      display: 'flex',
      paddingRight: '64px',
      flexDirection: 'column',
      '@media(max-width: 1359px)': {
        paddingRight: '32px'
      }
    },
    header: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      width: '100%',
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
      width: '100%'
    },
    main__wrap: {
      display: 'flex'
    },
    main__cards: {
      display: 'flex',
      alignItems: 'flex-start',
      flexWrap: 'wrap',
      width: '100%',
      maxWidth: '1080px'
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
