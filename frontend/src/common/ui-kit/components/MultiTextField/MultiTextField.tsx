import { ChangeEvent, FC, useCallback, useState } from 'react';
import Box from '@mui/material/Box';
import { styled } from '@mui/material/styles';

import { TextField } from '../index';
import { splitString } from './MultiTextField.utils';
import { PropsType, ValuesType } from './MultiTextField.types';

const FieldsContainer = styled(Box)({
  display: 'flex',
  gap: 12,
  flexWrap: 'wrap'
});

const NormalField = styled(TextField)({
  flex: 1,
  minWidth: 120
});

const WideField = styled(TextField)({
  flex: 2,
  minWidth: 200
});

const NarrowField = styled(TextField)({
  width: 100
});

const MultiTextField: FC<PropsType> = ({
  fieldName,
  delimiter,
  compoundValue,
  labels,
  fieldWidth = [],
  onFieldsChange
}) => {
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

  const getFieldComponent = (width?: string) => {
    switch (width) {
      case 'wide': return WideField;
      case 'narrow': return NarrowField;
      default: return NormalField;
    }
  };

  return (
    <Box>
      <FieldsContainer>
        {labels.map((label, index) => {
          const FieldComponent = getFieldComponent(fieldWidth[index]);
          return (
            <FieldComponent
              key={label}
              label={label}
              name={index.toString()}
              value={values[index]}
              onChange={onChangeHandler}
              error={!!errors[index]}
              helperText={errors[index]}
            />
          );
        })}
      </FieldsContainer>
    </Box>
  );
};

export default MultiTextField;
