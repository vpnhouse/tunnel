import { ChangeEvent, FC } from 'react';
import Collapse from '@mui/material/Collapse';
import MenuItem from '@mui/material/MenuItem';
import Select, { SelectChangeEvent } from '@mui/material/Select';
import Tooltip from '@mui/material/Tooltip';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import HelpOutlineRounded from '@mui/icons-material/HelpOutlineRounded';
import { styled } from '@mui/material/styles';

import { TextField } from '@common/ui-kit/components';
import { DomainConfig, Mode, ProxySchema } from '@common/components/DomainConfiguration/types';
import Switcher from '@common/ui-kit/components/Switcher';
import Checkbox from '@root/common/ui-kit/components/Checkbox';

import { FAQ_SSL_SERTIFICATE } from './constant';

const Root = styled(Box)({
  marginBottom: 16
});

const CheckboxWrapper = styled(Box)({
  display: 'flex',
  alignItems: 'center',
  marginBottom: 8
});

const StyledCheckbox = styled(Checkbox)({
  marginRight: 8
});

const StyledCollapse = styled(Collapse)({
  paddingLeft: 32
});

const StyledInput = styled(TextField)({
  marginBottom: 12
});

const Description = styled(Typography)(({ theme }) => ({
  fontSize: 14,
  color: theme.palette.text.secondary,
  marginBottom: 16
}));

const ProxyContainer = styled(Box)({
  display: 'flex',
  alignItems: 'flex-start',
  gap: 12,
  marginBottom: 12
});

const ProxySelector = styled(Select)({
  width: 120,
  '& .MuiSelect-icon': {
    color: 'inherit'
  }
});

const HelpIcon = styled(HelpOutlineRounded)(({ theme }) => ({
  marginLeft: 8,
  fontSize: 18,
  color: theme.palette.text.secondary,
  cursor: 'pointer'
}));

interface Props {
  domainConfig: DomainConfig;
  changeSettings: (event: ChangeEvent<HTMLElement>) => void;
  toggleIssueSSL: () => void;
  domainNameValidationError: string;
  withDomain: boolean;
  toggleWithDomain: () => void;
}

const DomainConfiguration: FC<Props> = ({ domainConfig, changeSettings, domainNameValidationError, toggleIssueSSL, withDomain, toggleWithDomain }) => {
  const { domain_name, issue_ssl, mode } = domainConfig;

  const handleSelectChange = (event: SelectChangeEvent<unknown>) => {
    changeSettings(event as unknown as ChangeEvent<HTMLElement>);
  };

  function renderDomainInput() {
    return (
      <StyledInput
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
    <Root>
      <CheckboxWrapper>
        <StyledCheckbox
          color="primary"
          id="domainName"
          checked={withDomain}
          onChange={toggleWithDomain}
        />
        <label htmlFor="domainName">I have a domain name</label>
      </CheckboxWrapper>

      <StyledCollapse in={withDomain}>
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
            <Description>
              The internet traffic is delivered directly to the instance, we can manage SSL certificates in that case. Use this mode if unsure.
            </Description>

            <CheckboxWrapper>
              <StyledCheckbox
                color="primary"
                id="issueSSL"
                checked={issue_ssl}
                onChange={toggleIssueSSL}
              />
              <label htmlFor="issueSSL">Issue SSL certificate</label>

              <Tooltip placement="right-start" title={FAQ_SSL_SERTIFICATE || ''}>
                <HelpIcon />
              </Tooltip>
            </CheckboxWrapper>
          </>
        )}

        {mode === Mode.ReverseProxy && (
          <>
            <ProxyContainer>
              <ProxySelector
                defaultValue={ProxySchema.https}
                onChange={handleSelectChange}
                name="schema"
                variant="outlined"
              >
                <MenuItem value={ProxySchema.https}>https://</MenuItem>
                <MenuItem value={ProxySchema.http}>http://</MenuItem>
              </ProxySelector>

              {renderDomainInput()}
            </ProxyContainer>
            <Description>
              The internet traffic is served by third-party software. SSL certificates should be managed by such software as well. Requires extra configuration of the web server software. For expert users.
            </Description>
          </>
        )}
      </StyledCollapse>
    </Root>
  );
};

export default DomainConfiguration;
