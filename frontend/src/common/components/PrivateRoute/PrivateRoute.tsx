import { FC, ReactNode } from 'react';
import { useUnit } from 'effector-react';
import { Navigate } from 'react-router-dom';

import { $authStore } from '@root/store/auth';
import { AUTH_ROUTE, INITIAL_CONFIGURATION } from '@constants/routes';
import { $initialSetup } from '@root/store/initialSetup';

interface PrivateRouteProps {
  children: ReactNode;
}

const PrivateRoute: FC<PrivateRouteProps> = ({ children }) => {
  const isAuthenticated = useUnit($authStore);
  const isInitialConfigurateDone = useUnit($initialSetup);

  if (!isInitialConfigurateDone) {
    return <Navigate to={INITIAL_CONFIGURATION} replace />;
  }

  return isAuthenticated ? <>{children}</> : <Navigate to={AUTH_ROUTE} replace />;
};

export default PrivateRoute;
