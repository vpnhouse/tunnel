import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    root: {
      width: '100%',
      height: 56
    },
    downloadLink: {
      color: palette.text.primary,
      textDecoration: 'none'
    }
  }));

export default useStyles;
