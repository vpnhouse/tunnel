import { ChangeEvent, FC, useCallback } from 'react';

import { TextField, Button } from '@common/ui-kit/components';
import DeleteIcon from '@root/common/assets/DeleteIcon';

import { PropsType } from './DnsRow.types';
import { INVALID_SYMBOLS, SYMBOL_ERRORS } from '../../index.constants';
import useStyles from './DnsRow.styles';

const DnsRow: FC<PropsType> = ({ dns, removeDnsFromList, onDnsChange, autoFocus, onBlur }) => {
  const classes = useStyles();

  const removeDnsHandler = useCallback(() =>
    removeDnsFromList(dns.id), [removeDnsFromList, dns.id]);

  const onChangeHandler = useCallback((event: ChangeEvent<HTMLInputElement>) => {
    const { value } = event.target;
    const regexp = INVALID_SYMBOLS.ipv4;
    const isInvalid = regexp.test(value);

    onDnsChange(dns.id, {
      dns: isInvalid ? dns.dns : value,
      error: isInvalid ? SYMBOL_ERRORS.dns : ''
    });
  }, [onDnsChange, dns.id, dns.dns]);

  return (
    <div className={classes.root}>
      <TextField
        className={classes.dnsField}
        fullWidth
        variant="outlined"
        label="DNS Address"
        type="text"
        value={dns.dns}
        error={!!dns.error}
        helperText={dns.error}
        onChange={onChangeHandler}
        onBlur={onBlur}
        autoFocus={autoFocus}
      />

      <Button
        className={classes.deleteButton}
        variant="contained"
        color="secondary"
        startIcon={<DeleteIcon />}
        onClick={removeDnsHandler}
      >
        Delete
      </Button>
    </div>
  );
};

export default DnsRow;
