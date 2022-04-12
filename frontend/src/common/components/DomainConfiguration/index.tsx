import React, { ChangeEvent, FC } from 'react';
import { Collapse, MenuItem, Select } from '@material-ui/core';
import clsx from 'clsx';

import { TextField } from '@common/ui-kit/components';
import { DomainConfig, Mode, ProxySchema } from '@common/components/DomainConfiguration/types';
import Switcher from '@common/ui-kit/components/Switcher';
import Checkbox from '@root/common/ui-kit/components/Checkbox';

import useStyles from './index.styles';

interface Props {
  domainConfig: DomainConfig;
  changeSettings: (event: ChangeEvent<HTMLElement>) => void;
  toggleIssueSSL: () => void;
  domainNameValidationError: string;
  withDomain: boolean;
  toggleWithDomain: () => void;
}

const DomainConfiguration: FC<Props> = ({ domainConfig, changeSettings, domainNameValidationError, toggleIssueSSL, withDomain, toggleWithDomain }) => {
  const classes = useStyles();

  const { domain_name, schema, issue_ssl, mode } = domainConfig;

  function renderDomainInput() {
    return (
      <TextField
        className={classes.input}
        fullWidth
        variant="outlined"
        label="Domain name"
        name="domain_name"
        placeholder="example.com"
        value={domain_name}
        error={!!domainNameValidationError}
        helperText={domainNameValidationError}
        onChange={changeSettings}
      />
    );
  }

  return (
    <div className={classes.root}>
      <div className={clsx(classes.checkboxWrapper, classes.domainToggler)}>
        <Checkbox
          color="primary"
          id="domainName"
          className={classes.checkbox}
          checked={withDomain}
          onChange={toggleWithDomain}
        />
        <label htmlFor="domainName">I have a domain name</label>
      </div>

      <Collapse in={withDomain} className={classes.collapse}>
        <Switcher
          name="mode"
          options={[Mode.Direct, Mode.ReverseProxy]}
          labels={['Domain Only', 'Reverse Proxy']}
          selected={mode}
          onChange={changeSettings}
        />

        {mode === Mode.Direct && (
          <>
            {renderDomainInput()}

            <div className={clsx(classes.checkboxWrapper, classes.issueSSL)}>
              <Checkbox
                color="primary"
                id="issueSSL"
                className={classes.checkbox}
                checked={issue_ssl}
                onChange={toggleIssueSSL}
              />
              <label htmlFor="issueSSL">Issue SSL certificate</label>
            </div>
          </>
        )}

        {mode === Mode.ReverseProxy && (
          <div className={classes.proxy}>
            <Select
              className={clsx(classes.input, classes.proxySelector)}
              value={schema}
              inputProps={{
                classes: {
                  icon: classes.proxySelectorIcon
                }
              }}
              // @ts-ignore ignored due to poor typing of MaterialUI selectors
              onChange={changeSettings}
              name="schema"
              variant="outlined"
            >
              <MenuItem value={ProxySchema.https}>https://</MenuItem>
              <MenuItem value={ProxySchema.http}>http://</MenuItem>
            </Select>

            {renderDomainInput()}
          </div>
        )}
      </Collapse>
    </div>
  );
};

export default DomainConfiguration;
