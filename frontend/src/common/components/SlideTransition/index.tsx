import React from 'react';
import { TransitionProps } from '@material-ui/core/transitions';
import { Slide } from '@material-ui/core';

const SlideTransition = React.forwardRef((
  props: TransitionProps,
  ref: React.Ref<unknown>
) => (
  <Slide
    appear
    {...props}
    ref={ref}
  />
));

export default SlideTransition;
