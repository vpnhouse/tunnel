import React from 'react';
import { SvgIcon, SvgIconProps } from '@material-ui/core';

const AddIcon = (props: SvgIconProps) => (
  <SvgIcon preserveAspectRatio="none" viewBox="0 0 14 14" {...props}>
    <path d="M7.75 0.75C7.75 0.335786 7.41421 0 7 0C6.58579 0 6.25 0.335786 6.25 0.75V6.25H0.75C0.335786 6.25 0 6.58579 0 7C0 7.41421 0.335786 7.75 0.75 7.75H6.25V13.25C6.25 13.6642 6.58579 14 7 14C7.41421 14 7.75 13.6642 7.75 13.25V7.75H13.25C13.6642 7.75 14 7.41421 14 7C14 6.58579 13.6642 6.25 13.25 6.25H7.75V0.75Z" fill="white" />
  </SvgIcon>
);

export default AddIcon;
