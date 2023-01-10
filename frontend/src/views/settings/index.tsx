import React, { ChangeEvent, FC, FormEvent, useCallback, useEffect, useMemo, useState } from 'react';
import { useStore } from 'effector-react';
import { Backdrop, CircularProgress, Typography, Paper, Tooltip } from '@material-ui/core';
import { HelpOutlineRounded } from '@material-ui/icons';

import { $settingsStore, getSettingsFx, changeSettingsFx } from '@root/store/settings';
import { $loadingStore, $statusStore } from '@root/store/status';
import { Button, TextField } from '@common/ui-kit/components';
import { VisibilityAdornment } from '@common/components';
import { SettingsRequest, SettingsResponseType, SettingsType } from '@root/store/settings/types';
import DomainConfiguration from '@common/components/DomainConfiguration';
import { DomainConfig, DomainEventTargetType, Mode } from '@common/components/DomainConfiguration/types';
import { DEFAULT_DOMAIN_CONFIG } from '@common/components/DomainConfiguration/constant';
import RefreshIcon from '@root/common/assets/RefreshIcon';
import SaveIcon from '@common/assets/SaveIcon';
import Checkbox from '@common/ui-kit/components/Checkbox';
import { MIMIMUM_PASSWORD_LENGTH } from '@constants/global';
import { getTruthStringLength } from '@common/utils/password';

import {
  NUMERIC_FIELDS,
  SYMBOL_SCHEMES,
  INVALID_SYMBOLS,
  SYMBOL_ERRORS,
  PATTERN_VALIDATION,
  PATTERN_ERRORS
} from './index.constants';
import { addIdtoDns, dnsNameValidation } from './index.utils';
import {
  DnsType,
  SettingsChangedType,
  SettingsErrorType,
  SettingsEventTargetType
} from './index.types';
import DnsSettings from './DnsSettings/DnsSettings';
import useStyles from './index.styles';


const Settings: FC = () => {
  const savedSettings: SettingsResponseType | null = useStore($settingsStore);
  const { restart_required } = useStore($statusStore);
  const isLoading = useStore($loadingStore);
  const classes = useStyles();

  const [settings, setSettings] = useState<SettingsType | null>(savedSettings);
  const [domainConfig, setDomainConfig] = useState<DomainConfig>(savedSettings?.domain || DEFAULT_DOMAIN_CONFIG);
  const [withDomain, setWithDomain] = useState(!!savedSettings?.domain);

  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [dnsList, setDnsList] = useState<DnsType[]>([]);
  const [validationError, setValidationError] = useState<SettingsErrorType>({} as SettingsErrorType);

  useEffect(() => {
    setSettings(savedSettings);
    if (savedSettings?.dns) setDnsList(addIdtoDns(savedSettings.dns));
    if (savedSettings?.domain) {
      setDomainConfig(savedSettings.domain);
      setWithDomain(true);
    }
    setValidationError({} as SettingsErrorType);
  }, [savedSettings]);

  useEffect(() => {
    getSettingsFx();
  }, []);

  const refreshPageHandler = useCallback(() => {
    getSettingsFx();
  }, []);

  const changeSettingsHandler = useCallback((event: ChangeEvent<HTMLElement>) => {
    const { name, value } = event.target as SettingsEventTargetType;

    const invalidSymbols = SYMBOL_SCHEMES[name];
    const isInvalid = invalidSymbols ? INVALID_SYMBOLS[invalidSymbols].test(value) : false;

    setValidationError((prevError) => ({
      ...prevError,
      [name]: isInvalid ? SYMBOL_ERRORS[invalidSymbols!] : ''
    }));

    setSettings((prevSettings: SettingsType | null) => ({
      ...prevSettings,
      /** If field contains invalid symbols do not change it */
      [name]: isInvalid
        ? (prevSettings?.[name] || '')
        /** Some fields must contain numeric data */
        : (NUMERIC_FIELDS.includes(name))
          ? Number(value)
          : value
    }) as SettingsType);
  }, []);

  const changeDnsHandler = useCallback((newDns: DnsType[]) => {
    setDnsList(() => newDns);
    setSettings((prevSettings: SettingsType | null) => ({
      ...prevSettings,
      /** Not empty DNS add to settings */
      dns: newDns.filter((item) => item.dns).map((item) => item.dns)
    }) as SettingsType);
  }, []);

  const settingsChanged: SettingsChangedType | null = useMemo(() => {
    if (!settings) return null;

    const { dns, admin_password: password, confirm_password: cPassword, ...rest } = settings;
    const { domain } = savedSettings as SettingsResponseType;
    const dnsChanged = JSON.stringify(dns) !== JSON.stringify(savedSettings?.dns);

    let domainChanged = withDomain !== !!domain;

    if (domain && (
      domain.domain_name !== domainConfig.domain_name
      || domain.mode !== domainConfig.mode
      || domain.schema !== domainConfig.schema
      || domain.issue_ssl !== domainConfig.issue_ssl
    )) {
      domainChanged = true;
    }

    const restFieldsChanged = (Object.entries(rest) as Entries<SettingsResponseType>)
      .reduce<SettingsChangedType>((changesList, [key, value]) => ({
        ...changesList,
        [key]: value !== savedSettings?.[key]
      }), {} as SettingsChangedType);

    return {
      ...restFieldsChanged,
      dns: dnsChanged,
      domain: domainChanged,
      admin_password: password !== ''
    };
  }, [settings, savedSettings, domainConfig, withDomain]);

  const isSettingsChanged = useMemo(() => {
    if (!settingsChanged) return false;

    return Object.values(settingsChanged).some((isFieldChanged) => isFieldChanged);
  }, [settingsChanged]);

  const validateDns = useCallback(() => {
    const checkedDns = dnsList.map((dns) => ({
      ...dns,
      error: PATTERN_VALIDATION.dns?.(dns.dns)
    }));
    const isValid = checkedDns.every((dns) => !dns.error);

    !isValid && setDnsList(() => checkedDns as DnsType[]);

    return isValid;
  }, [dnsList]);

  const validate = useCallback(() => {
    const { dns, domain, ...rest } = settingsChanged as SettingsChangedType;

    /** If DNS was changed, check if all DNS fields are in valid format */
    const idDnsValid = dns ? validateDns() : true;
    const domainNameError = domain && withDomain ? dnsNameValidation(domainConfig.domain_name) : '';

    domainNameError && setValidationError((prevError) => ({
      ...prevError,
      domain_name: domainNameError
    }));

    const passwordLengthOk = settingsChanged?.admin_password
      ? (settings?.admin_password && getTruthStringLength(settings.admin_password) >= MIMIMUM_PASSWORD_LENGTH)
      : true;

    /** Check if password and its confirmation match */
    const passwordsMatch = settingsChanged?.admin_password
      ? settings?.admin_password === settings?.confirm_password
      : true;

    !passwordsMatch && setValidationError((prevError) => ({
      ...prevError,
      confirm_password: PATTERN_ERRORS.passwordNotMatch
    }));

    !passwordLengthOk && setValidationError((prevError) => ({
      ...prevError,
      admin_password: PATTERN_ERRORS.passwordLength
    }));

    /** Check if other changed fields are in valid format */
    const restFieldsErrors = (Object.entries(rest) as Entries<Omit<SettingsChangedType, 'domain'>>)
      .filter(([_, value]) => value)
      .reduce<SettingsErrorType>((errors, [field, _]) => ({
        ...errors,
        [field]: PATTERN_VALIDATION[field] ? PATTERN_VALIDATION[field]?.(settings?.[field] as never) : ''
      }), {} as SettingsErrorType);
    const isRestFieldsValid = Object.values(restFieldsErrors).every((error) => !error);

    !isRestFieldsValid && setValidationError((prevError) => ({
      ...prevError,
      ...restFieldsErrors
    }));

    return idDnsValid && passwordsMatch && isRestFieldsValid && !domainNameError && passwordLengthOk;
  }, [settings, settingsChanged, validateDns, domainConfig, withDomain]);

  const saveChangesHandler = useCallback((e: FormEvent<HTMLFormElement> | React.MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation();
    e.preventDefault();

    if (!validate()) return;

    const { confirm_password, admin_password, ...rest } = settings as SettingsType;

    const body: SettingsRequest = {
      ...rest,
      admin_password: admin_password || undefined,
      domain: null
    };

    if (withDomain) {
      const { domain_name, mode, issue_ssl, schema } = domainConfig;
      if (domainConfig.mode === Mode.Direct) {
        body.domain = {
          mode,
          domain_name,
          issue_ssl
        };
      } else {
        body.domain = {
          mode,
          domain_name,
          schema
        };
      }
    }

    changeSettingsFx(body);
  }, [settings, validate, withDomain, domainConfig]);

  const toggleShowPasswordHandler = useCallback(() =>
    setShowPassword((prevState) => !prevState), []);

  const toggleShowConfirmPasswordHandler = useCallback(() =>
    setShowConfirmPassword((prevState) => !prevState), []);

  function changeDomainConfig(event: ChangeEvent<HTMLElement>) {
    const { name, value } = event.target as DomainEventTargetType;

    const invalidSymbols = SYMBOL_SCHEMES[name];
    const isInvalid = invalidSymbols ? INVALID_SYMBOLS[invalidSymbols].test(value) : false;

    setDomainConfig((prevState) => ({
      ...prevState,
      [name]: isInvalid ? prevState[name] : value
    } as DomainConfig));
  }

  function toggleIssueSSL() {
    setDomainConfig({
      ...domainConfig,
      issue_ssl: !domainConfig?.issue_ssl
    } as DomainConfig);
  }

  function toggleSendStats() {
    if (settings) {
      setSettings({
        ...settings,
        send_stats: !settings.send_stats
      });
    }
  }

  return (
    <div className={classes.root}>
      <div className={classes.header}>
        <Typography variant="h1" color="textPrimary">
          Settings
        </Typography>
        <div className={classes.buttonLine}>
          <Button
            className={classes.resetButton}
            variant="contained"
            color="secondary"
            onClick={refreshPageHandler}
            startIcon={<RefreshIcon />}
          >
            Reset
          </Button>
          <Button
            className={classes.saveButton}
            disabled={!isSettingsChanged && !restart_required}
            variant="contained"
            color="primary"
            onClick={saveChangesHandler}
            startIcon={<SaveIcon />}
          >
            Save & Reload
          </Button>
        </div>
      </div>
      <div className={classes.settings}>
        <Backdrop className={classes.backdrop} open={isLoading}>
          <Paper className={classes.backdropPaper}>
            <CircularProgress />
            <Typography variant="subtitle1">
              Settings are saved
            </Typography>
            <Typography variant="subtitle1">
              Service is reloading
            </Typography>
          </Paper>

        </Backdrop>
        {/* The custom value of the autoComplete prop is based on Chrome behavior */}
        {/* More here: https://stackoverflow.com/questions/30053167 */}
        <form onSubmit={saveChangesHandler} autoComplete="nofill">
          <div className={classes.settingsBlock}>
            <Typography variant="h4">Create new password</Typography>

            <TextField
              fullWidth
              variant="outlined"
              label="New password"
              type={showPassword ? 'text' : 'password'}
              name="admin_password"
              value={settings?.admin_password || ''}
              error={!!validationError?.admin_password}
              helperText={validationError?.admin_password || ''}
              onChange={changeSettingsHandler}
              autoComplete="false"
              endAdornment={(
                <VisibilityAdornment
                  tabIndex="-1"
                  showPassword={showPassword}
                  toggleShowPasswordHandler={toggleShowPasswordHandler}
                />
              )}
            />
            <TextField
              fullWidth
              variant="outlined"
              label="Confirm new password"
              type={showConfirmPassword ? 'text' : 'password'}
              name="confirm_password"
              value={settings?.confirm_password || ''}
              error={!!validationError?.confirm_password}
              helperText={validationError?.confirm_password || ''}
              onChange={changeSettingsHandler}
              autoComplete="off"
              endAdornment={(
                <VisibilityAdornment
                  tabIndex="-1"
                  showPassword={showConfirmPassword}
                  toggleShowPasswordHandler={toggleShowConfirmPasswordHandler}
                />
              )}
            />
          </div>

          <div className={classes.settingsBlock}>
            <Typography variant="h4">Server</Typography>

            <TextField
              className={classes.publicKey}
              disabled
              fullWidth
              variant="outlined"
              label="Wireguard Public Key"
              name="wireguard_public_key"
              value={settings?.wireguard_public_key || ''}
              InputProps={{
                readOnly: true
              }}
            />

            <TextField
              fullWidth
              variant="outlined"
              label="Public IPv4 address"
              name="wireguard_server_ipv4"
              value={settings?.wireguard_server_ipv4 || ''}
              error={!!validationError?.wireguard_server_ipv4}
              helperText={validationError?.wireguard_server_ipv4 || ''}
              onChange={changeSettingsHandler}
            />

            <TextField
              fullWidth
              variant="outlined"
              label="Internal network subnet"
              name="wireguard_subnet"
              value={settings?.wireguard_subnet || ''}
              error={!!validationError?.wireguard_subnet}
              helperText={validationError?.wireguard_subnet || ''}
              onChange={changeSettingsHandler}
            />
          </div>

          <div className={classes.settingsBlock}>
            <TextField
              fullWidth
              variant="outlined"
              label="Wireguard Access Port"
              name="wireguard_server_port"
              value={settings?.wireguard_server_port || ''}
              error={!!validationError?.wireguard_server_port}
              helperText={validationError?.wireguard_server_port || ''}
              onChange={changeSettingsHandler}
            />
          </div>

          <div className={classes.settingsBlock}>
            <Typography variant="h4">Domain</Typography>

            <DomainConfiguration
              domainConfig={domainConfig}
              changeSettings={changeDomainConfig}
              domainNameValidationError={validationError.domain_name}
              toggleIssueSSL={toggleIssueSSL}
              withDomain={withDomain}
              toggleWithDomain={() => setWithDomain(!withDomain)}
            />
          </div>

          <div className={classes.settingsBlock}>
            <Typography variant="h4">DNS Servers</Typography>

            <DnsSettings
              dns={dnsList}
              changeDnsHandler={changeDnsHandler}
            />
          </div>

          <div className={classes.checkboxWrapper}>
            <Checkbox
              color="primary"
              id="sendStats"
              className={classes.checkbox}
              checked={settings ? settings.send_stats : false}
              onChange={toggleSendStats}
            />
            <label htmlFor="sendStats">Enable statistics</label>

            <Tooltip placement="right-start" title="We are a small open source team, so we are gathering statistics to make our product better.">
              <HelpOutlineRounded className={classes.field__faq_icon} />
            </Tooltip>
          </div>

          <Button type="submit" className={classes.hidden} />
        </form>
      </div>
    </div>
  );
};

export default Settings;
