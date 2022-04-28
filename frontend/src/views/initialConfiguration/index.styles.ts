import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette, zIndex, typography }) =>
  createStyles({
    root: {
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      fontFamily: typography.fontFamily,
      width: '100%',
      background: palette.background.default
    },
    container: {
      width: 'max-content',
      margin: '0 auto'
    },
    header: {
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      width: '700px',
      margin: '36px 0',
      '-webkit-user-select': 'none',
      '-moz-user-select': 'none',
      '-ms-user-select': 'none',
      userSelect: 'none'
    },
    settings: {
      height: '100%',
      overflow: 'auto',
      width: 930,
      color: palette.text.primary
    },
    buttonLine: {
      display: 'flex',
      justifyContent: 'flex-end',
      '& > :not(:first-child)': {
        marginLeft: '12px'
      },
      ' & button': {
        padding: '0 28px'
      }
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
      marginBottom: 8
    },
    checkbox: {
      padding: 0,
      paddingRight: 9
    },
    field__faq_wrap: {
      display: 'flex',
      alignItems: 'center',
      width: '100%'
    },
    field__faq_icon: {
      marginLeft: '12px',
      cursor: 'pointer',
      opacity: 0.5
    }
  }));

export default useStyles;
