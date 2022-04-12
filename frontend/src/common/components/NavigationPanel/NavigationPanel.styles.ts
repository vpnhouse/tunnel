import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette, breakpoints }) =>
  createStyles({
    root: {
      height: '100%',
      width: 240,
      boxSizing: 'border-box',
      backgroundColor: palette.background.paper,
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      marginRight: 64,
      padding: '64px 32px 64px',
      [breakpoints.down('md')]: {
        padding: '0 48px'
      }
    },
    logo: {
      width: 130,
      height: 32,
      marginBottom: 64
    }
  }));

export default useStyles;
