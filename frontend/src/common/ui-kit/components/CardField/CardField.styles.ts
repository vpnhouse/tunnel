import { createStyles, makeStyles } from '@material-ui/core/styles';

import { StylesPropsTipe } from './CardField.types';

const useStyles = makeStyles(({ palette }) =>
  createStyles({
    areaRoot: {
      display: 'flex',
      justifyContent: 'space-between'
    },
    textBlock: ({ tableView }: StylesPropsTipe) => ({
      marginBottom: '12px',
      display: tableView ? 'flex' : 'block',
      alignItems: tableView ? 'baseline' : 'unset'
    }),
    actions: {
      display: 'flex',
      flexDirection: 'column',
      margin: '8px -10px 0 8px'
    },
    disable__control: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'flex-end',
      minWidth: 160,
      marginLeft: 14,
      '&>label': {
        marginRight: 0
      },
      '@media(max-width: 480px)': {
        marginLeft: 0
      }
    },
    field__withControl: {
      display: 'flex',
      alignItems: 'center',
      width: '100%',
      marginTop: 8,
      marginBottom: 4,
      '@media(max-width: 480px)': {
        flexDirection: 'column',
        alignItems: 'flex-start'
      }
    },
    caption: ({ tableView }: StylesPropsTipe) => ({
      display: 'flex',
      alignItems: 'end',
      marginBottom: tableView ? 0 : '4px',
      '& svg': {
        display: 'block',
        height: '15px',
        marginRight: '5px'
      },
      width: tableView ? '100px' : 'unset'
    }),
    value: {
      flex: '1 0 auto'
    },
    error: {
      color: palette.error.main,
      paddingTop: '5px'
    },
    dateTimePicker: {
      '@media(max-width: 480px)': {
        display: 'flex',
        flexDirection: 'column',
        '& > *': {
          width: '100% !important'
        }
      }
    }
  }));

export default useStyles;
