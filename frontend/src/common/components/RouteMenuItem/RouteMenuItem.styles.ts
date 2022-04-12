import { createStyles, makeStyles } from '@material-ui/core/styles';

import { SelectedProps } from './RouteMenuItem.types';

const useStyles = makeStyles(({ palette, typography }) =>
  createStyles({
    itemRoot: {
      borderRadius: 12,
      paddingLeft: 16,
      marginBottom: 16,
      color: palette.text.secondary,
      transition: 'background-color 0.2s ease',
      '&.Mui-selected': {
        backgroundColor: palette.background.default,
        color: palette.primary.main,
        '&:hover': {
          backgroundColor: palette.background.default
        }
      },
      '&:hover': {
        '&:not(.Mui-selected)': {
          backgroundColor: palette.background.paper,
          color: palette.text.primary,
          '& path': {
            fill: palette.text.primary
          }
        }
      }
    },
    itemSelected: {
      backgroundColor: palette.background.default
    },
    listItemIconRoot: ({ selected }: SelectedProps) => ({
      minWidth: 36,
      '& path': {
        fill: selected ? palette.primary.main : palette.text.secondary
      }
    }),
    iconRoot: {
      fontSize: '20px',
      height: 24,
      width: 24
    },
    itemTextRoot: {
      display: 'flex',
      justifyContent: 'space-between'
    },
    primaryText: {
      ...typography.subtitle1,
      display: 'inline'
    },
    secondaryText: ({ selected }: SelectedProps) => ({
      ...typography.subtitle1,
      color: selected ? palette.primary.main : palette.text.primary
    })
  }));

export default useStyles;
