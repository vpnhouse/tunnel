export const BYTES_MEASURE_LIMITS = [
  {
    limit: 1000,
    label: 'B'
  },
  {
    limit: 1000000,
    label: 'KB'
  },
  {
    limit: 1000000000,
    label: 'MB'
  },
  {
    limit: 1000000000000,
    label: 'GB'
  },
  {
    limit: 1000000000000000,
    label: 'TB'
  }
];

export const FETCH_STATS_INTERVAL = 5000;

export const TABS = [
  {
    value: 'all',
    label: 'All'
  },
  {
    value: 'wireguard',
    label: 'Wireguard'
  },
  {
    value: 'iprose',
    label: 'IPRose'
  }
];
