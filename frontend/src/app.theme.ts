import { createMuiTheme } from '@material-ui/core/styles';

const theme = createMuiTheme({
  shape: {
    borderRadius: 6
  },
  palette: {
    common: {
      black: '#121212',
      white: '#FBFBFB'
    },
    background: {
      default: '#0F121A',
      paper: '#181E2E'
    },
    action: {
      disabled: 'rgba(43, 49, 66, 0.4)',
      disabledBackground: '#9B99AC'
    },
    /** green */
    primary: {
      main: '#1FC477',
      light: '#1EE57B',
      dark: '#1E9463'
    },
    /** blue */
    secondary: {
      main: 'rgba(0, 0, 0, 0)',
      light: 'rgba(31, 196, 119, 0.1)',
      dark: '#284362'
    },
    error: {
      main: '#EA4F4F',
      light: '#FC7070',
      dark: '#5B1414'
    },
    warning: {
      main: '#F99A43'
    },
    info: {
      main: '#233D5A'
    },
    text: {
      primary: '#FBFBFB',
      secondary: 'rgba(251, 251, 251, 0.6)',
      disabled: 'rgba(251, 251, 251, 0.2)',
      hint: 'rgba(251, 251, 251, 0.4)'
    }
  },
  typography: {
    fontFamily: "'Ubuntu', 'Roboto', 'Helvetica', 'Arial', sans-serif",
    button: {
      fontFamily: "'Ubuntu', 'Roboto', 'Helvetica', 'Arial', sans-serif",
      fontWeight: 500,
      fontSize: '16px',
      lineHeight: '24px',
      textTransform: 'none'
    },
    h1: {
      fontFamily: "'Ubuntu', 'Roboto', 'Helvetica', 'Arial', sans-serif",
      fontWeight: 400,
      fontSize: '48px',
      lineHeight: '56px',
      letterSpacing: '0.0075em',
      '@media(max-width: 991px)': {
        fontSize: '36px',
        lineHeight: '36px'
      }
    },
    h2: {
      fontFamily: "'Ubuntu', 'Roboto', 'Helvetica', 'Arial', sans-serif",
      fontWeight: 500,
      fontSize: '24px',
      lineHeight: '32px'
    },
    h4: {
      fontFamily: "'Ubuntu', 'Roboto', 'Helvetica', 'Arial', sans-serif",
      fontWeight: 700,
      fontSize: '16px',
      lineHeight: '24px'
    },
    h5: {
      fontFamily: "'Ubuntu', 'Roboto', 'Helvetica', 'Arial', sans-serif",
      fontWeight: 700,
      fontSize: '20px',
      lineHeight: '24px'
    },
    h6: {
      fontFamily: "'Ubuntu', 'Roboto', 'Helvetica', 'Arial', sans-serif",
      fontWeight: 400,
      fontSize: '23px',
      lineHeight: '24px'
    },
    subtitle1: {
      fontFamily: "'Ubuntu', 'Roboto', 'Helvetica', 'Arial', sans-serif",
      fontWeight: 400,
      fontSize: '16px',
      lineHeight: '24px',
      textTransform: 'none'
    },
    body1: {
      fontFamily: "'Ubuntu', 'Roboto', 'Helvetica', 'Arial', sans-serif",
      fontWeight: 400,
      fontSize: '16px',
      lineHeight: '21px'
    },
    caption: {
      fontFamily: "'Ubuntu', 'Roboto', 'Helvetica', 'Arial', sans-serif",
      fontWeight: 400,
      fontSize: '12px',
      lineHeight: '18px'
    }
  },
  breakpoints: {
    values: {
      xs: 0,
      sm: 600,
      md: 1280,
      lg: 1440,
      xl: 1920
    }
  },
  props: {
    MuiButtonBase: {
      disableRipple: true,
      disableTouchRipple: true
    }
  }
});

export {
  theme
};
