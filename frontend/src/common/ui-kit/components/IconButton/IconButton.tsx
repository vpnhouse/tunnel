import React, { FC } from 'react';
import { Icon, IconButton as MaterialIconButton, Tooltip } from '@material-ui/core';
import clsx from 'clsx';

import { PropsType } from './IconButton.types';
import useStyles from './IconButton.styles';
import { DEFAULT_ICON_PROPS } from './IconButton.constants';

const IconButton: FC<PropsType> = ({
  color,
  onClick,
  icon,
  title = '',
  iconProps = DEFAULT_ICON_PROPS,
  className,
  tabIndex
}) => {
  const classes = useStyles(iconProps);

  return (
    <Tooltip
      title={title}
      placement="right"
      classes={{
        tooltip: classes.tooltip
      }}
    >
      <MaterialIconButton
        className={clsx(classes.root, classes[color], className)}
        onClick={onClick}
        tabIndex={tabIndex}
      >
        <div className={classes.iconWrapper} id="iconButton">
          <Icon
            component={icon}
            classes={{
              root: classes.iconRoot
            }}
          />
        </div>
      </MaterialIconButton>
    </Tooltip>
  );
};

export default IconButton;
