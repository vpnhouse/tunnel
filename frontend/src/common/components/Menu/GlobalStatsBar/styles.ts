import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette, typography }) =>
  createStyles({
    root: {
      borderRadius: 12,
      padding: 16,
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
    }
  }));

export default useStyles;
