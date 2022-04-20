import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette, zIndex, typography }) =>
  createStyles({
    root: {
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
      fontFamily: typography.fontFamily,
      paddingRight: '64px',
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
    resetButton: {
      width: 115
    },
    saveButton: {
      width: 177
    },
    settings: {
      height: '100%',
      overflow: 'auto',
      width: '100%',
      color: palette.text.primary
    },
    settingsBlock: {
      width: 320,
      marginBottom: 32,
      '&>div:not(:last-child)': {
        margin: 0,
        marginBottom: 12
      },
      '&>h4': {
        marginBottom: 16
      }
    },
    publicKey: {
      width: 480
    },
    buttonLine: {
      display: 'flex',
      justifyContent: 'flex-end',
      '& > :not(:first-child)': {
        marginLeft: '12px'
      },
      ' & button': {
        padding: 20,
        '& svg': {
          height: 16,
          width: 16,
          marginRight: 4
        }
      }
    },
    hidden: {
      display: 'none'
    },
    backdrop: {
      zIndex: zIndex.drawer + 1,
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      backgroundColor: `${palette.common.black}99` // 60% opacity
    },
    backdropPaper: {
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      padding: '24px 48px',
      backgroundColor: `${palette.background.paper}CC` // 80% opacity
    },
    checkboxWrapper: {
      display: 'flex',
      alignItems: 'center',
      marginBottom: 32
    },
    checkbox: {
      padding: 0,
      paddingRight: 9
    },
    field__faq_icon: {
      marginLeft: '12px',
      cursor: 'pointer',
      opacity: 0.5
    }
  }));

export default useStyles;
