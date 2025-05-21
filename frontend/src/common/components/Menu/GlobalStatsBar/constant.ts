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
    value: 'stats_global',
    label: 'All'
  },
  {
    value: 'stats_wireguard',
    label: 'Wireguard'
  },
  {
    value: 'stats_iprose',
    label: 'IPRose'
  },
  {
    value: 'stats_proxy',
    label: 'Proxy'
  }
];
