import { createStyles, makeStyles } from '@material-ui/core/styles';

import { IconPropsType } from './IconButton.types';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    root: {
      width: 36,
      height: 36,

      '&:hover': {
        cursor: 'pointer',
        '& div': {
          '&:before': {
            visibility: 'visible'
          }
        }
      },
      '&:active': {
        '& div': {
          '&:before': {
            backgroundColor: 'rgba(0, 0, 0, 0.6)'
          }
        }
      }
    },
    primary: {
      color: palette.common.white,
      padding: 0,
      '&:hover': {
        backgroundColor: 'none',
        color: palette.primary.main
      }
    },
    error: {
      color: palette.error.main,
      padding: '5px',
      height: '34px',
      '&:hover': {
        backgroundColor: 'none',
        color: palette.error.light
      }
    },
    tooltip: {
      backgroundColor: palette.secondary.main,
      fontFamily: "'Roboto', 'Helvetica', 'Arial', sans-serif",
      fontSize: '12px',
      fontWeight: 400
    },
    iconWrapper: {
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      position: 'relative',
      zIndex: 1,

      '&:before': {
        zIndex: -1,
        content: '" "',
        display: 'block',
        visibility: 'hidden',
        width: 36,
        height: 36,
        borderRadius: '50%',
        position: 'absolute',
        backgroundColor: '#2B3142'
      }
    },
    iconRoot: ({ fontSize } : IconPropsType) => ({
      fontSize,
      width: 12,
      height: 12
    })
  }));

export default useStyles;
