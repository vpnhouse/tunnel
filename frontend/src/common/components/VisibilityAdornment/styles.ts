import { makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) => ({
  root: {
    padding: 0,
    '&:hover': {
      '& path': {
        fill: palette.text.secondary
      }
    }
  },
  disabled: {
    '& path': {
      fill: palette.text.hint
    }
  }
}));

export default useStyles;
