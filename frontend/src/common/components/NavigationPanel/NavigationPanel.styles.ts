import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) =>
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
      '@media(max-height: 640px)': {
        overflowY: 'auto'
      },
      '@media(max-width: 1359px)': {
        marginRight: 32
      },
      '@media(max-width: 991px)': {
        width: 64,
        padding: '32px 8px 32px'
      }
    },
    logo: {
      width: 130,
      height: 32,
      marginBottom: 64,
      '@media(max-width: 991px)': {
        display: 'none'
      }
    },
    logo__mobile: {
      display: 'none',
      '@media(max-width: 991px)': {
        display: 'block',
        height: 32,
        marginBottom: 64
      }
    }
  }));

export default useStyles;
