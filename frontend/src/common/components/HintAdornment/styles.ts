import { makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) => ({
  root: {
    padding: 0,
    height: 16,
    width: 16,
    '&:hover': {
      color: palette.text.secondary
    }
  }
}));

export default useStyles;
