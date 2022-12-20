import React, { useCallback, useEffect, useState } from 'react';

import { fetchData } from '@root/store/utils';
import { GLOBAL_STATS } from '@constants/apiPaths';
import { GlobalStats, GlobalStatsResponse } from '@common/components/Menu/GlobalStatsBar/types';
import { BYTES_MEASURE_LIMITS } from '@common/components/Menu/GlobalStatsBar/constant';

import useStyles from './styles';

const FETCH_STATS_INTERVAL = 5000;

const GlobalStatsBar = () => {
  const classes = useStyles();

  const [stats, setStats] = useState<GlobalStats | null>(null);

  const convertBytes = useCallback((bytes: number): string => {
    // eslint-disable-next-line no-restricted-syntax
    for (const [index, metric] of BYTES_MEASURE_LIMITS.entries()) {
      const { limit, label } = metric;

      if (bytes < limit) {
        return `${(bytes / (1000 ** index)).toFixed(2)} ${label}`;
      }
    }

    return `${bytes} B`;
  }, []);

  const fetchStats = useCallback(async () => {
    const response = await fetchData(GLOBAL_STATS);
    const data: GlobalStatsResponse = await response.json();

    setStats({
      ...data,
      traffic_up: convertBytes(data.traffic_up),
      traffic_down: convertBytes(data.traffic_down),
      traffic_up_speed: convertBytes(data.traffic_up_speed) + `PS`,
      traffic_down_speed: convertBytes(data.traffic_down_speed) + `PS`
    });
  }, [convertBytes]);

  useEffect(() => {
    fetchStats();
    const fetchStatsInterval = setInterval(fetchStats, FETCH_STATS_INTERVAL);

    return () => {
      clearInterval(fetchStatsInterval);
    };
  }, [fetchStats]);

  if (!stats) {
    return null;
  }

  const { peers_active_1h, peers_active_1d, peers_connected, peers_total, traffic_up, traffic_down, traffic_up_speed, traffic_down_speed} = stats;

  return (
    <div className={classes.root}>
      <h3 className={classes.title}>Global stats</h3>

      <div className={classes.row}>
        <span>Total peers:</span>

        <span>{peers_total}</span>
      </div>

      <div className={classes.row}>
        <span>Peers connected:</span>

        <span>{peers_connected}</span>
      </div>

      <div className={classes.row}>
        <span>Peers active last hour:</span>

        <span>{peers_active_1h}</span>
      </div>

      <div className={classes.row}>
        <span>Peers active last day:</span>

        <span>{peers_active_1d}</span>
      </div>

      <div className={classes.row}>
        <span>Upstream traffic:</span>

        <span>{traffic_up}</span>
      </div>

      <div className={classes.row}>
        <span>Downstream traffic:</span>

        <span>{traffic_down}</span>
      </div>

      <div className={classes.row}>
        <span>Upstream speed:</span>

        <span>{traffic_up_speed}</span>
      </div>

      <div className={classes.row}>
        <span>Downstream speed:</span>

        <span>{traffic_down_speed}</span>
      </div>

    </div>
  );
};

export default GlobalStatsBar;
