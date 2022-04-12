import { makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(({ palette }) => ({
  root: {
    display: 'flex',
    fontSize: '16px',
    width: 260,
    marginBottom: 12,
    backgroundColor: '#2B3142',
    borderRadius: 8,
    padding: 4,
    '&>input': {
      display: 'none'
    }
  },
  switcherOption: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    padding: '8px 12px',
    height: 24,
    cursor: 'pointer',
    position: 'relative',
    borderRadius: 8,
    flexGrow: 1,
    textTransform: 'capitalize',
    fontWeight: 400,
    fontSize: 16,
    lineHeight: '14px',
    fontFamily: 'Ubuntu',
    transition: 'background-color 0.5s ease',
    '&:last-child': {
      marginLeft: '-5px'
    }
  },
  switcherOptionActive: {
    backgroundColor: palette.primary.main,
    cursor: 'unset',
    zIndex: 2
  },
  switcherOptionDisabled: {
    '&:hover': {
      backgroundColor: '#3B3F63'
    },
    '&>label': {
      cursor: 'pointer'
    }
  }
}));

export default useStyles;
