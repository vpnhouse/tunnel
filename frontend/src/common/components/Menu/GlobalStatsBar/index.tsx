import React, { useEffect, useState } from 'react';

import { fetchData } from '@root/store/utils';
import { GLOBAL_STATS } from '@constants/apiPaths';
import { GlobalStats } from '@common/components/Menu/GlobalStatsBar/types';

import useStyles from './styles';

const FETCH_STATS_INTERVAL = 15000;

const GlobalStatsBar = () => {
  const classes = useStyles();

  const [stats, setStats] = useState<GlobalStats | null>(null);

  async function fetchStats() {
    const response = await fetchData(GLOBAL_STATS);
    const data = await response.json();

    setStats(data);
  }

  useEffect(() => {
    fetchStats();
    const fetchStatsInterval = setInterval(fetchStats, FETCH_STATS_INTERVAL);

    return () => {
      clearInterval(fetchStatsInterval);
    };
  }, []);

  if (!stats) {
    return null;
  }

  const { peers_active, peers_connected, peers_total, traffic_tx, traffic_rx } = stats;

  return (
    <div className={classes.root}>
      <h3 className={classes.title}>Global stats</h3>

      <div className={classes.row}>
        <span>Total peers:</span>

        <span>{peers_total}</span>
      </div>

      <div className={classes.row}>
        <span>Peers during session:</span>

        <span>{peers_connected}</span>
      </div>

      <div className={classes.row}>
        <span>Active peers:</span>

        <span>{peers_active}</span>
      </div>

      <div className={classes.row}>
        <span>Upstream traffic:</span>

        <span>{traffic_tx}</span>
      </div>

      <div className={classes.row}>
        <span>Downstream traffic:</span>

        <span>{traffic_rx}</span>
      </div>
    </div>
  );
};

export default GlobalStatsBar;
