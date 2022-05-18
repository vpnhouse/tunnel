// This util is created to properly count length of a password that contains emoji
// since "❤️".length === 2, this checker is required in such cases
// for more details check out https://blog.jonnew.com/posts/poo-dot-length-equals-two
export function getTruthStringLength(str: string) {
  const joiner = '\u{200D}';
  const split = str.split(joiner);
  let count = 0;

  split.forEach((s) => {
    // removing the variation selectors
    const num = Array.from(s.split(/[\ufe00-\ufe0f]/).join('')).length;
    count += num;
  });

  // assuming the joiners are used appropriately
  return count / split.length;
}

export function getAsterisksFromString(str: string) {
  return new Array(getTruthStringLength(str)).fill('*').join('');
}
