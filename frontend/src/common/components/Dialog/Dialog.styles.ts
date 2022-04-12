import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ typography, palette }) =>
  createStyles({
    paper: {
      padding: '20px 28px',
      minWidth: '400px'
    },
    title: {
      ...typography.h5,
      padding: 0,
      marginBottom: '12px'
    },
    content: {
      padding: 0
    },
    contentText: {
      ...typography.subtitle1,
      color: palette.text.primary,
      marginBottom: '32px'
    },
    actions: {
      padding: 0,
      '& > :not(:first-child)': {
        marginLeft: '12px'
      }
    }
  }));

export default useStyles;
