import { createStyles, makeStyles } from '@material-ui/core/styles';

import { StylesPropsType } from './TextArea.types';

const useStyles = makeStyles(() =>
  createStyles({
    root: ({ tableView }: StylesPropsType) => ({
      display: 'flex',
      justifyContent: 'space-between',
      width: tableView ? '536px' : 'unset'
    }),
    textArea: ({ allVisible }: StylesPropsType) => ({
      whiteSpace: 'pre-line',
      display: '-webkit-box',
      '-webkit-line-clamp': 3,
      '-webkit-box-orient': allVisible ? 'unset' : 'vertical',
      overflow: 'hidden',
      overflowWrap: 'anywhere',
      height: '100%',
      fontSize: '14px'
    }),
    actions: {
      display: 'flex',
      flexDirection: 'column',
      marginLeft: '8px',
      marginRight: '-10px'
    }
  }));

export default useStyles;
