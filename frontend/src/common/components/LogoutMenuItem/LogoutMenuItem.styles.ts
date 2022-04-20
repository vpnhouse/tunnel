import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette, typography }) =>
  createStyles({
    itemRoot: {
      borderRadius: '12px',
      paddingLeft: '22px',
      position: 'absolute',
      left: 0,
      bottom: 0,
      color: palette.text.secondary,
      '&:hover': {
        backgroundColor: palette.background.paper,
        color: palette.text.primary,
        '& path': {
          fill: palette.text.primary
        }
      },
      '@media(max-width: 991px)': {
        padding: 8,
        alignItems: 'center',
        justifyContent: 'center'
      }
    },
    itemSelected: {
      backgroundColor: palette.background.default
    },
    listItemIconRoot: {
      minWidth: '32px',
      '& path': {
        fill: palette.text.secondary
      }
    },
    iconRoot: {
      '& svg': {
        height: '20px',
        width: '20px'
      }
    },
    primaryText: {
      ...typography.subtitle1,
      display: 'inline',
      '@media(max-width: 991px)': {
        display: 'none'
      }
    }
  }));

export default useStyles;
