import React, { ChangeEvent, FC, useCallback, useState } from 'react';

import { TextField } from '../index';
import { splitString } from './MultiTextField.utils';
import { PropsType, ValuesType } from './MultiTextField.types';
import useStyles from './MultiTextField.styles';

const MultiTextField: FC<PropsType> = ({
  fieldName,
  delimiter,
  compoundValue,
  labels,
  fieldWidth = [],
  onFieldsChange
}) => {
  const classes = useStyles();
  const [values, setValues] = useState<ValuesType>(splitString(compoundValue, delimiter, labels.length));
  const [errors, setErrors] = useState<ValuesType>(splitString('', delimiter, labels.length));

  const onChangeHandler = useCallback((e: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;

    /** Fields except last should not contain delimiter */
    const regexp = new RegExp(delimiter);
    const isInvalid = (Number(name) < Object.keys(values).length - 1) && regexp.test(value);
    const updatedValues = {
      ...values,
      [name]: isInvalid ? values[name] : value
    };

    setErrors((prevError) => ({
      ...prevError,
      [name]: isInvalid ? `The field cannot contain symbol ${delimiter}` : ''
    }));
    setValues(updatedValues);

    onFieldsChange(fieldName, Object.values(updatedValues).join(delimiter));
  }, [values, delimiter, fieldName, onFieldsChange]);

  return (
    <div>
      <div className={classes.root}>
        {labels.map((label, index) => (
          <TextField
            className={classes[fieldWidth[index] || 'normal']}
            key={label}
            label={label}
            name={index.toString()}
            value={values[index]}
            onChange={onChangeHandler}
            error={!!errors[index]}
            helperText={errors[index]}
          />
        ))}
      </div>
    </div>

  );
};

export default MultiTextField;
