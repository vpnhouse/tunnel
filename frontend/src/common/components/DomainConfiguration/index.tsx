import React, { ChangeEvent, FC } from 'react';
import { Collapse, MenuItem, Select, Tooltip } from '@material-ui/core';
import clsx from 'clsx';
import { HelpOutlineRounded } from '@material-ui/icons';

import { TextField } from '@common/ui-kit/components';
import { DomainConfig, Mode, ProxySchema } from '@common/components/DomainConfiguration/types';
import Switcher from '@common/ui-kit/components/Switcher';
import Checkbox from '@root/common/ui-kit/components/Checkbox';

import useStyles from './index.styles';
import { FAQ_SSL_SERTIFICATE } from './constant';

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

  const { domain_name, issue_ssl, mode } = domainConfig;

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
            <p className={classes.text_description}>
              The internet traffic is delivered directly to the instance, we can manage SSL certificates in that case. Use this mode if unsure.
            </p>

            <div className={clsx(classes.checkboxWrapper, classes.issueSSL)}>
              <Checkbox
                color="primary"
                id="issueSSL"
                className={classes.checkbox}
                checked={issue_ssl}
                onChange={toggleIssueSSL}
              />
              <label htmlFor="issueSSL">Issue SSL certificate</label>

              <Tooltip placement="right-start" title={FAQ_SSL_SERTIFICATE || ''}>
                <HelpOutlineRounded className={classes.field__faq_icon} />
              </Tooltip>
            </div>
          </>
        )}

        {mode === Mode.ReverseProxy && (
          <>
            <div className={classes.proxy}>
              <Select
                defaultValue={ProxySchema.https}
                className={clsx(classes.input, classes.proxySelector)}
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
            <p className={classes.text_description}>
              The internet traffic is served by third-party software. SSL certificates should be managed by such software as well. Requires extra configuration of the web server software. For expert users.
            </p>
          </>
        )}
      </Collapse>
    </div>
  );
};

export default DomainConfiguration;
