import React, { FC } from 'react';
import { ListItem, ListItemIcon, ListItemText, SvgIcon } from '@material-ui/core';
import { Link as RouterLink } from 'react-router-dom';

import useStyles from './RouteMenuItem.styles';
import { PropsType } from './RouteMenuItem.types';

const RouteMenuItem: FC<PropsType> = ({
  selected,
  icon,
  route,
  pageTitle,
  extraInfo,
  onClick
}) => {
  const classes = useStyles({ selected });

  return (
    <ListItem
      button
      component={RouterLink}
      to={route}
      selected={selected}
      classes={{
        root: classes.itemRoot,
        selected: classes.itemSelected
      }}
      onClick={onClick}
    >
      <ListItemIcon
        classes={{
          root: classes.listItemIconRoot
        }}
      >
        <SvgIcon
          classes={{
            root: classes.iconRoot
          }}
          component={icon}
        />
      </ListItemIcon>
      <ListItemText
        classes={{
          root: classes.itemTextRoot,
          primary: classes.primaryText,
          secondary: classes.secondaryText
        }}
        primary={pageTitle}
        secondary={extraInfo}
      />
    </ListItem>
  );
};

export default RouteMenuItem;
