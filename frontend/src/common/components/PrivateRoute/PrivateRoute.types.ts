import { ComponentType } from 'react';
import { RouteComponentProps, RouteProps } from 'react-router';

export type PropsType = RouteProps & {
  component: ComponentType<RouteComponentProps<any>> | ComponentType<any>
}
