import React from 'react';
import { TransitionProps } from '@material-ui/core/transitions';
import { Slide } from '@material-ui/core';

const PeerModalTransition = React.forwardRef((
  props: TransitionProps & {
    children: React.ReactElement<any, any>;
  },
  ref: React.Ref<unknown>
) => (
  <Slide
    appear
    ref={ref}
    {...props}
  />
));

export default PeerModalTransition;
