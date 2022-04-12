import { INITIAL_CONFIGURATION } from '@constants/routes';

import { $initialSetup, checkConfigurationFx, setInitialSetupFx, setInitialSetupState } from './index';
import { addNotification } from '../notifications';
import { checkToken } from '../auth';
import { setGlobalLoading } from '../globalLoading';

$initialSetup.on(setInitialSetupState, (_, res) => res);

checkConfigurationFx.doneData.watch(() => {
  setInitialSetupState(true);
  checkToken();
});

checkConfigurationFx.failData.watch((error) => {
  if (error?.status !== 409) {
    setInitialSetupState(true);
  } else {
    window.location.pathname !== INITIAL_CONFIGURATION && window.location.replace(INITIAL_CONFIGURATION);
  }
});

checkConfigurationFx.finally.watch(() => setTimeout(() => setGlobalLoading(false), 1000));

setInitialSetupFx.doneData.watch(() => {
  setInitialSetupState(true);
  checkToken();
});

setInitialSetupFx.finally.watch(() => setTimeout(() => setGlobalLoading(false), 1000));

setInitialSetupFx.failData.watch(
  (error) => error.json().then((err) => {
    addNotification({
      type: 'error',
      prefix: 'serverError',
      message: err.error
    });
  })
);
