import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ typography }) =>
  createStyles({
    root: {
      fontFamily: typography.fontFamily
    },
    domainToggler: {
      marginBottom: 12
    },
    collapse: {
      transition: 'all 0.4s ease-in-out !important',
      opacity: 1,
      '&.MuiCollapse-hidden': {
        opacity: 0
      }
    },
    issueSSL: {
      marginTop: 12
    },
    checkboxWrapper: {
      display: 'flex',
      alignItems: 'center'
    },
    checkbox: {
      padding: 0,
      paddingRight: 9
    },
    input: {
      height: 56,
      margin: 0
    },
    proxy: {
      display: 'flex',
      alignItems: 'center'
    },
    proxySelector: {
      backgroundColor: '#2B3142',
      maxWidth: 120,
      minWidth: 120,
      marginRight: 12,
      '& fieldset': {
        border: 'none'
      },
      '&:hover': {
        backgroundColor: '#3B3F63'
      }
    },
    proxySelectorIcon: {
      fill: 'white'
    },
    field__faq_icon: {
      marginLeft: '12px',
      cursor: 'pointer',
      opacity: 0.5
    },
    text_description: {
      fontSize: '12px',
      lineHeight: '16px',
      color: '#fff',
      opacity: 0.7
    }
  }));

export default useStyles;
