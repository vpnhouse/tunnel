import { createStyles, makeStyles } from '@material-ui/core/styles';

import { StylesPropsType } from './DatePicker.types';

const useStyles = makeStyles(({ palette, typography }) =>
  createStyles({
    root: {
      display: 'block',
      width: '48%',
      height: 56,
      borderRadius: 8
    },
    labelRoot: ({ isEmpty }: StylesPropsType) => ({
      color: isEmpty ? palette.text.hint : palette.text.primary
    }),
    inputLabelFilled: {
      '&.MuiInputLabel-marginDense': {
        transform: 'translate(14px, 18px) scale(1)'
      },
      '&.MuiInputLabel-shrink': {
        transform: 'translate(14px, 6px) scale(0.75)'
      },
      '&.Mui-error': {
        color: palette.text.hint
      }
    },
    notchedOutline: {
      border: 'none'
    },
    adornedEnd: ({ isEmpty }: StylesPropsType) => ({
      color: isEmpty ? palette.text.hint : palette.text.primary,
      paddingRight: 0,
      '&.Mui-focused': {
        color: palette.primary.main
      },
      '&.Mui-error': {
        color: palette.text.secondary
      },
      '& button': {
        color: 'inherit'
      }
    }),
    inputRoot: {
      height: 56,
      backgroundColor: '#2B3142',
      '&:hover': {
        backgroundColor: '#3B3F63'
      },
      '&.Mui-error': {
        backgroundColor: palette.error.dark
      }
    },
    input: {
      color: palette.text.primary,
      paddingBottom: 0,
      '&:-webkit-autofill': {
        '-webkit-text-fill-color': palette.text.primary,
        '-webkit-box-shadow': `0 0 0px 1000px ${palette.background.paper} inset`
      },
      '&.Mui-disabled': {
        color: palette.text.disabled
      }
    },
    inputMarginDense: {
      paddingTop: '4px',
      paddingBottom: '4px',
      width: '100%'
    },
    helperText: {
      textAlign: 'end'
    },
    paper: {
      '& .MuiPickersCalendarHeader-iconButton': {
        backgroundColor: palette.secondary.main,
        color: palette.common.white
      },
      '& .MuiPickersCalendarHeader-transitionContainer p': {
        ...typography.subtitle1
      }
    }
  }));

export default useStyles;
