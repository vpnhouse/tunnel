import { createStyles, makeStyles } from '@material-ui/core/styles';
import { OutlinedTextFieldProps } from '@material-ui/core/TextField/TextField';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    root: {
      display: 'block',
      minHeight: 56
    },
    labelRoot: {
      color: palette.text.hint,
      top: 2,
      left: 4
    },
    inputLabelFilled: {
      '&.MuiInputLabel': {
        top: 10
      },
      '&.Mui-error': {
        color: palette.text.hint
      }
    },
    adornedEnd: ({ value }: Partial<OutlinedTextFieldProps>) => ({
      color: value ? '' : palette.text.hint,
      cursor: 'pointer',
      paddingRight: 20,
      '&.Mui-error': {
        color: palette.error.main
      }
    }),
    inputRoot: {
      borderRadius: 8,
      paddingLeft: 4,
      caretColor: palette.primary.main,
      backgroundColor: '#2B3142',
      '&:hover': {
        backgroundColor: '#3B3F63',
        '&.Mui-disabled': {
          backgroundColor: palette.action.disabled
        }
      },
      '&.Mui-disabled': {
        backgroundColor: palette.action.disabled
      },
      '&.Mui-focused': {
        backgroundColor: '#3B3F63'
      },
      '&.Mui-error': {
        backgroundColor: palette.error.dark,
        caretColor: 'unset'
      }
    },
    input: {
      color: palette.text.primary,
      '&:-webkit-autofill': {
        '-webkit-text-fill-color': palette.text.primary,
        transition: 'background-color 5000s ease-in-out 0s'
      },
      '&.Mui-disabled': {
        color: palette.text.disabled
      }
    },
    inputMarginDense: {
      paddingTop: '4px',
      paddingBottom: '4px'
    },
    helperText: {
      textAlign: 'end'
    }
  }));

export default useStyles;
