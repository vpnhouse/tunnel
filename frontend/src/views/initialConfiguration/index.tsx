import React, { ChangeEvent, FormEvent, useCallback, useState } from 'react';
import { Backdrop, CircularProgress, Paper, Typography } from '@material-ui/core';
import { useStore } from 'effector-react';
import { Autorenew } from '@material-ui/icons';
import { Redirect } from 'react-router';

import { Button, TextField } from '@common/ui-kit/components';
import { VisibilityAdornment } from '@root/common/components';
import { $loadingStore } from '@root/store/status';
import { InitialSetupData, InitialSetupDomain } from '@root/store/initialSetup/types';
import { $initialSetup, setInitialSetupFx } from '@root/store/initialSetup';
import { setGlobalLoading } from '@root/store/globalLoading';
import DomainConfiguration from '@common/components/DomainConfiguration';
import { Mode, ProxySchema } from '@root/common/components/DomainConfiguration/types';
import Checkbox from '@common/ui-kit/components/Checkbox';

import { INVALID_SYMBOLS, PATTERN_ERRORS, SYMBOL_ERRORS, SYMBOL_SCHEMES } from '../settings/index.constants';
import useStyles from './index.styles';
import { Config, ConfigTargetType, PasswordError } from './types';
import { dnsNameValidation } from '../settings/index.utils';
import { checkRequiredFields, generateSubMaskValue } from './utils';


const InitialConfiguration = () => {
  const classes = useStyles();
  const isInitialConfigurateDone = useStore($initialSetup);
  const isLoading = useStore($loadingStore);

  const [withDomain, setWidthDomain] = useState(false);
  const [sendStats, setSendStats] = useState(true);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [validationError, setValidationError] = useState<PasswordError>({} as PasswordError);
  const [settings, setSettings] = useState<Config>({
    admin_password: '',
    confirm_password: '',
    domain_name: '',
    mode: Mode.Direct,
    issue_ssl: false,
    schema: ProxySchema.https,
    wireguard_subnet: generateSubMaskValue()
  });


  const toggleShowPasswordHandler = useCallback(() => setShowPassword((prevState) => !prevState), []);
  const toggleShowConfirmPasswordHandler = useCallback(() => setShowConfirmPassword((prevState) => !prevState), []);

  const changeSettingsHandler = useCallback((event: ChangeEvent<HTMLElement>) => {
    const { name, value } = event.target as ConfigTargetType;

    const check = INVALID_SYMBOLS[SYMBOL_SCHEMES[name]];

    let isInvalid = false;

    if (check) {
      isInvalid = check.test(value);
    }

    setValidationError((prevError) => ({
      ...prevError,
      [name]: isInvalid ? SYMBOL_ERRORS[SYMBOL_SCHEMES[name]] : ''
    }));

    if (name === 'mode') {
      setSettings((prevSettings) => ({
        ...prevSettings,
        [name]: isInvalid ? (prevSettings?.[name] || '') : value as Mode,
        domain_name: '',
        issue_ssl: false,
        schema: ProxySchema.https
      }));
    } else {
      setSettings((prevSettings) => ({
        ...prevSettings,
        [name]: isInvalid ? (prevSettings?.[name] || '') : value
      }));
    }
  }, []);

  const validate = useCallback(() => {
    const validateRequiredFields = checkRequiredFields(settings);
    const passwordsMatch = settings?.admin_password === settings?.confirm_password;
    const domainNameError = withDomain ? dnsNameValidation(settings.domain_name) : '';
    const isAllFieldsValid = Object.values(validateRequiredFields).every((error) => !error);

    const errors = {
      ...(!passwordsMatch ? { confirm_password: PATTERN_ERRORS.password } : {}),
      ...(domainNameError ? { domain_name: PATTERN_ERRORS.dnsName } : {}),
      ...validateRequiredFields
    };

    if (!isAllFieldsValid || domainNameError || !passwordsMatch) {
      setValidationError((prevState) => ({
        ...prevState,
        ...errors
      }));

      return false;
    }

    return true;
  }, [withDomain, settings]);

  const resetState = useCallback(() => {
    setSettings((prevState) => ({
      ...prevState,
      admin_password: '',
      confirm_password: '',
      domain_name: '',
      mode: Mode.Direct,
      schema: ProxySchema.https
    }));
    setWidthDomain(false);
    setValidationError({} as PasswordError);
  }, []);

  const save = useCallback((e: FormEvent<HTMLFormElement> | React.MouseEvent<HTMLButtonElement>) => {
    e.preventDefault();
    e.stopPropagation();

    if (!validate()) {
      return;
    }
    const data: InitialSetupData = {
      admin_password: settings.admin_password,
      server_ip_mask: settings.wireguard_subnet,
      send_stats: sendStats
    };

    if (withDomain) {
      const domainData: InitialSetupDomain = {
        domain_name: settings.domain_name,
        mode: settings.mode
      };

      if (settings.mode === Mode.Direct) {
        domainData.issue_ssl = settings.issue_ssl;
      } else {
        domainData.schema = settings.schema;
      }

      data.domain = domainData;
    }

    setInitialSetupFx(data).then(() => setGlobalLoading(true));
  }, [withDomain, settings, validate, sendStats]);

  function toggleIssueSSL() {
    setSettings({
      ...settings,
      issue_ssl: !settings.issue_ssl
    });
  }

  function toggleSendStats() {
    setSendStats(!sendStats);
  }

  return (
    isInitialConfigurateDone ? (
      <Redirect to="/" />
    ) : (
      <section className={classes.root}>
        <form onSubmit={save} className={classes.container}>
          <div className={classes.header}>
            <Typography variant="h1" color="textPrimary">
              Initial Configuration
            </Typography>
            <div className={classes.buttonLine}>
              <Button
                variant="contained"
                type="button"
                color="secondary"
                onClick={resetState}
                startIcon={<Autorenew />}
              >
                Reset
              </Button>
              <Button
                variant="contained"
                type="submit"
                color="primary"
              >
                Save
              </Button>
            </div>
          </div>
          <div className={classes.settings}>
            <Backdrop className={classes.backdrop} open={isLoading}>
              <Paper className={classes.backdropPaper}>
                <CircularProgress />
                <Typography variant="subtitle1">
                  Configuration are saved
                </Typography>
                <Typography variant="subtitle1">
                  Service is reloading
                </Typography>
              </Paper>
            </Backdrop>
            <TextField
              fullWidth
              variant="outlined"
              label="Password"
              type={showPassword ? 'text' : 'password'}
              name="admin_password"
              value={settings?.admin_password}
              error={!!validationError?.admin_password}
              helperText={validationError?.admin_password || ''}
              onChange={changeSettingsHandler}
              endAdornment={(
                <VisibilityAdornment
                  showPassword={showPassword}
                  toggleShowPasswordHandler={toggleShowPasswordHandler}
                  tabIndex="-1"
                />
              )}
            />
            <TextField
              fullWidth
              variant="outlined"
              label="Confirm Password"
              type={showConfirmPassword ? 'text' : 'password'}
              name="confirm_password"
              value={settings?.confirm_password}
              error={!!validationError?.confirm_password}
              helperText={validationError?.confirm_password || ''}
              onChange={changeSettingsHandler}
              endAdornment={(
                <VisibilityAdornment
                  showPassword={showConfirmPassword}
                  toggleShowPasswordHandler={toggleShowConfirmPasswordHandler}
                  tabIndex="-1"
                />
              )}
            />
            <TextField
              fullWidth
              variant="outlined"
              label="Subnet mask"
              name="wireguard_subnet"
              error={!!validationError?.wireguard_subnet}
              helperText={validationError?.wireguard_subnet || ''}
              onChange={changeSettingsHandler}
              value={settings?.wireguard_subnet}
              style={{ marginBottom: 8 }}
            />

            <div className={classes.checkboxWrapper}>
              <Checkbox
                color="primary"
                id="sendStats"
                className={classes.checkbox}
                checked={sendStats}
                onChange={toggleSendStats}
              />
              <label htmlFor="sendStats">Count my registration</label>
            </div>

            <DomainConfiguration
              domainConfig={settings}
              changeSettings={changeSettingsHandler}
              domainNameValidationError={validationError.domain_name}
              toggleIssueSSL={toggleIssueSSL}
              withDomain={withDomain}
              toggleWithDomain={() => setWidthDomain(!withDomain)}
            />
          </div>
        </form>
      </section>
    )
  );
};

export default InitialConfiguration;
