import { FC, useCallback, useState, MouseEvent } from 'react';
import { v4 as uuid } from 'uuid';

import { TextField } from '@common/ui-kit/components';

import { PropsType, DnsDataType } from './DnsSettings.types';
import useStyles from './DnsSettings.styles';
import DnsRow from './DnsRow/DnsRow';

const DnsSettings: FC<PropsType> = ({ dns = [], changeDnsHandler }) => {
  const classes = useStyles();

  const [focusOnLast, setFocusOnLast] = useState(false);

  function blurFocus() {
    setFocusOnLast(false);
  }

  function preventFocus(e: MouseEvent) {
    e.preventDefault();
  }

  const addDnsHandler = useCallback(() => {
    changeDnsHandler([
      ...dns,
      {
        id: uuid(),
        dns: '',
        error: ''
      }
    ]);

    setFocusOnLast(true);
  }, [changeDnsHandler, dns]);

  const removeDnsHandler = useCallback((id: string) =>
    changeDnsHandler(dns.filter((item) => item.id !== id)), [changeDnsHandler, dns]);

  const onChangeHandler = useCallback((id: string, dnsData: DnsDataType) =>
    changeDnsHandler(dns.map((item) =>
      (item.id === id
        ? {
          ...item,
          ...dnsData
        }
        : item))), [changeDnsHandler, dns]);

  return (
    <div className={classes.dnsBlock}>
      {dns.map((item, index) => (
        <DnsRow
          key={item.id}
          dns={item}
          removeDnsFromList={removeDnsHandler}
          onDnsChange={onChangeHandler}
          onBlur={blurFocus}
          autoFocus={focusOnLast && index === dns.length - 1}
        />
      ))}

      <TextField
        className={classes.addNewField}
        variant="outlined"
        onClick={addDnsHandler}
        value=""
        label="Add new DNS"
        onMouseDown={preventFocus}
      />
    </div>
  );
};

export default DnsSettings;
