import * as React from 'react';
import { ChangeEvent, FC } from 'react';
import clsx from 'clsx';

import useStyles from './styles';

interface Props {
  options: [string, string];
  selected: string;
  onChange: (event: ChangeEvent<HTMLElement>) => void;
  labels?: [string, string];
  name?: string;
}

const Switcher: FC<Props> = ({ options, selected, onChange, name, labels }) => {
  const classes = useStyles();
  const [firstOption, secondOption] = options;

  return (
    <div className={classes.root}>
      <input type="radio" value={firstOption} name={name || 'switcher'} id="switcherOption1" checked={selected === firstOption} onChange={onChange} />
      <label
        className={clsx(classes.switcherOption, selected === firstOption ? classes.switcherOptionActive : classes.switcherOptionDisabled)}
        htmlFor="switcherOption1"
      >
        {labels ? labels[0] : firstOption}
      </label>

      <input type="radio" value={secondOption} name={name || 'switcher'} id="switcherOption2" checked={selected === secondOption} onChange={onChange} />
      <label
        className={clsx(classes.switcherOption, selected === secondOption ? classes.switcherOptionActive : classes.switcherOptionDisabled)}
        htmlFor="switcherOption2"
      >
        {labels ? labels[1] : secondOption}
      </label>
    </div>
  );
};

export default Switcher;
