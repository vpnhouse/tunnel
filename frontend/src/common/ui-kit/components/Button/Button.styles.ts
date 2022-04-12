import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette, typography }) =>
  createStyles({
    root: {
      height: '56px',
      padding: '0 42px',
      color: palette.text.primary,
      boxShadow: 'none',
      borderRadius: 8,
      transition: 'background-color 0 ease',
      '&:hover': {
        boxShadow: 'none',
        '& path': {
          fill: palette.text.primary
        }
      },
      '&.Mui-disabled': {
        backgroundColor: palette.action.disabled,
        color: palette.text.hint,
        '& path': {
          fill: palette.text.secondary
        }
      },
      '& path': {
        fill: palette.text.primary
      }
    },
    containedPrimary: {
      '&:hover': {
        backgroundColor: palette.primary.light
      },
      '&:focus': {
        backgroundColor: palette.primary.light
      }
    },
    containedSecondary: {
      '&:hover': {
        backgroundColor: palette.secondary.light
      }
    },
    text: {
      ...typography.subtitle1,
      fill: palette.primary.main,
      '&:hover': {
        backgroundColor: palette.background.paper,
        color: palette.primary.light,
        fill: palette.primary.light
      }
    },
    label: {
      color: 'inherit'
    }
  }));

export default useStyles;
