import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    caption: {
      display: 'flex',
      alignItems: 'end',
      '& svg': {
        display: 'block',
        height: '15px',
        marginRight: '5px'
      },
      width: '100px'
    },
    error: {
      color: palette.error.main,
      paddingTop: '5px'
    },
    textBlock: {
      marginBottom: '16px',
      display: 'flex',
      alignItems: 'baseline'
    },
    value: {
      flex: '1 0 auto'
    },
    header: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      height: '24px',
      marginBottom: '16px'
    },
    actions: {
      display: 'flex',
      flexDirection: 'column',
      margin: '14px -10px 0 11px'
    },
    leftLine: {
      minWidth: '245px',
      borderTop: `1px solid ${palette.text.secondary}`
    },
    rightLine: {
      width: '100%',
      borderTop: `1px solid ${palette.text.secondary}`
    },
    fieldLine: {
      display: 'flex',
      alignItems: 'center'
    },
    inputLine: {
      display: 'flex',
      alignItems: 'flex-start'
    },
    text: {
      overflowWrap: 'anywhere'
    }
  }));

export default useStyles;
