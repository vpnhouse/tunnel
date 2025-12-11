import { forwardRef, Ref, ReactElement } from 'react';
import { TransitionProps } from '@mui/material/transitions';
import Slide from '@mui/material/Slide';

const SlideTransition = forwardRef(function SlideTransition(
  props: TransitionProps & { children: ReactElement },
  ref: Ref<unknown>
) {
  return (
    <Slide
      appear
      direction="up"
      ref={ref}
      {...props}
    />
  );
});

export default SlideTransition;
