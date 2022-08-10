import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ typography, palette }) =>
  createStyles({
    paper: {
      padding: 32,
      maxWidth: 770,
      borderRadius: 12
    },
    title: {
      ...typography.h5,
      padding: 0,
      fontWeight: 500,
      marginBottom: 24,
      fontSize: 24,
      lineHeight: '32px'
    },
    closeDialog: {
      height: 12,
      width: 12,
      position: 'absolute',
      top: 42,
      right: 42
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
    },
    buttons: {
      display: 'flex',
      justifyContent: 'space-between',
      width: '100%',
      marginTop: 24
    },
    downloadLink: {
      color: palette.text.primary,
      textDecoration: 'none'
    }
  }));

export default useStyles;
