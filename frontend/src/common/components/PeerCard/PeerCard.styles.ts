import { createStyles, makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    paper: {
      width: '100%',
      maxWidth: '528px',
      padding: '24px 32px 24px',
      margin: '0 12px 12px 0',
      position: 'relative',
      boxSizing: 'border-box',
      '&:hover': {
        '& #peerCardActions': {
          visibility: 'visible'
        }
      }
    },
    paperError: {
      backgroundColor: palette.error.dark
    },
    dialog: {
      padding: '32px 32px 24px',
      position: 'relative'
    },
    closeDialog: {
      height: 12,
      width: 12,
      position: 'absolute',
      top: 34,
      right: 42
    },
    title: {
      marginBottom: 24
    },
    form: {
      '& > div:last-child': {
        marginBottom: 0
      }
    },
    actions: {
      position: 'absolute',
      right: 20,
      top: 70,
      visibility: 'hidden'
    },
    deleteIcon: {
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      position: 'relative',
      zIndex: 1,
      width: 56,
      height: 56,

      '& svg': {
        height: 16,
        width: 16
      },

      '&:before': {
        zIndex: -1,
        content: '" "',
        display: 'block',
        visibility: 'hidden',
        width: 56,
        height: 56,
        borderRadius: '50%',
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: '#2B3142'
      },
      '&:hover': {
        cursor: 'pointer',
        '&:before': {
          visibility: 'visible'
        }
      },
      '&:active': {
        '&:before': {
          backgroundColor: 'rgba(0, 0, 0, 0.6)'
        }
      }
    },
    publicKey: {
      width: '601px',
      height: '70px'
    },
    ipv4: {
      width: '305px',
      height: '70px'
    },
    claims: {
      marginBottom: '28px',
      '& .MuiOutlinedInput-inputMultiline': {
        height: '100%',
        paddingRight: '1px'
      }
    },
    buttonLine: {
      display: 'flex',
      justifyContent: 'flex-end',
      marginTop: 12,
      '& > :not(:last-child)': {
        marginRight: '16px'
      },
      '& > button': {
        width: 83,
        boxShadow: 'none'
      }
    },
    commonError: {
      display: 'flex',
      alignItems: 'end',
      padding: '8px 0 12px',
      '& svg': {
        display: 'block',
        height: '15px',
        marginRight: '5px'
      }
    }
  }));

export default useStyles;
