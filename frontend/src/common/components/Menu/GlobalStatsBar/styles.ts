import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette, typography }) =>
  createStyles({
    root: {
      borderRadius: 12,
      paddingBottom: 16,
      backgroundColor: `${palette.background.default}60`,
      color: palette.text.primary,
      fontFamily: typography.fontFamily,
      display: 'flex',
      flexDirection: 'column',
      marginTop: 48,
      '@media(max-width: 991px)': {
        display: 'none'
      }
    },
    title: {
      fontSize: 12,
      marginBottom: 20,
      marginTop: 0,
      color: palette.primary.main
    },
    row: {
      fontSize: 10,
      display: 'flex',
      justifyContent: 'space-between',
      '&:not(:last-child)': {
        marginBottom: 8
      }
    },
    tabsContainer: {
      marginBottom: 16
    },
    tabs: {
      minHeight: 32,
      '& .MuiTab-root': {
        minHeight: 32,
        padding: '6px 12px',
        fontSize: 12
      },
      '& .MuiButtonBase-root': {
        minWidth: '0px'
      }
    },
    statsContent: {
      marginTop: 8,
      padding: '0px 16px'
    }
  }));

export default useStyles;
