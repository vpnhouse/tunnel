import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette, typography }) =>
  createStyles({
    root: {
      display: 'flex',
      alignItems: 'center',
      height: '48px',
      padding: '0 12px',
      color: palette.text.primary,
      '&:hover': {
        backgroundColor: palette.background.paper,
        color: palette.primary.light,
        cursor: 'pointer'
      }
    },
    label: {
      ...typography.subtitle1,
      paddingLeft: '8px',
      whiteSpace: 'nowrap'
    }
  }));

export default useStyles;
