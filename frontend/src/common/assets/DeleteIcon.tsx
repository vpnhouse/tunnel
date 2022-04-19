import React from 'react';
import { SvgIcon, SvgIconProps } from '@material-ui/core';

const DeleteIcon = (props: SvgIconProps) => (
  <SvgIcon preserveAspectRatio="none" viewBox="0 0 16 16" {...props}>
    <path
      fillRule="evenodd"
      clipRule="evenodd"
      d="M7.75 1.5C7.05962 1.5 6.5 2.05964 6.5 2.75V3H9.5V2.75C9.5 2.05964 8.94038 1.5 8.25 1.5H7.75ZM11 3V2.75C11 1.23122 9.76882 0 8.25 0H7.75C6.23118 0 5 1.23122 5 2.75V3H2.7509H1C0.585786 3 0.25 3.33579 0.25 3.75C0.25 4.16421 0.585786 4.5 1 4.5H2.06243L2.84401 13.4883C2.96759 14.9094 4.15725 16 5.58367 16H10.4164C11.8428 16 13.0325 14.9094 13.156 13.4883L13.156 13.4883L13.9377 4.5H15C15.4142 4.5 15.75 4.16421 15.75 3.75C15.75 3.33579 15.4142 3 15 3H13.2492H11ZM12.432 4.5H3.56809L4.33838 13.3584C4.39454 14.0043 4.93528 14.5 5.58367 14.5H10.4164C11.0648 14.5 11.6055 14.0043 11.6617 13.3584L11.6617 13.3584L12.432 4.5Z"
      fill="white"
    />
  </SvgIcon>
);

export default DeleteIcon;