export const splitString = (initialString: string, delimiter: string, length: number) => {
  const subStrings = initialString.split(delimiter);
  const valuesArr = initialString
    ? [
      ...subStrings.slice(0, length - 1),
      subStrings.slice(length - 1).join(delimiter)
    ]
    : new Array(length).fill('');

  return valuesArr.reduce((list, value, index) => ({
    ...list,
    [index]: value
  }), {});
};
