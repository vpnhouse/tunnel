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
    field__faq: {
      display: 'flex',
      justifyContent: 'center',
      flexDirection: 'column'
    },
    disable__control: {
      '& .Mui-checked': {
        color: '#1FC477'
      },
      '& .MuiSwitch-colorSecondary + .MuiSwitch-track': {
        backgroundColor: '#FBFBFB'
      },
      '& .MuiSwitch-colorSecondary.Mui-checked + .MuiSwitch-track': {
        backgroundColor: '#fff',
        opacity: '0.9'
      }
    },
    field__faq_wrap: {
      display: 'flex',
      alignItems: 'center',
      width: '100%'
    },
    field__faq_icon: {
      marginLeft: '12px',
      cursor: 'pointer',
      opacity: 0.5
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
    }
  }));

export default useStyles;
