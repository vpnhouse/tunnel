import React from 'react';

import useStyles from './index.styles';

const GlobalLoader = () => {
  const classes = useStyles();

  return (
    <section className={classes.section}>
      <h3 className={classes.title}>Loading...</h3>
    </section>
  );
};

export default GlobalLoader;
