import React, { useCallback, useEffect, useState } from 'react';
import { Tabs, Tab } from '@material-ui/core';

import { components } from '@schema';
import { fetchData } from '@root/store/utils';
import { GLOBAL_STATS } from '@constants/apiPaths';
import { BYTES_MEASURE_LIMITS, FETCH_STATS_INTERVAL, TABS } from '@common/components/Menu/GlobalStatsBar/constant';

import useStyles from './styles';

type TabType = 'stats_global' | 'stats_iprose' | 'stats_proxy' | 'stats_wireguard';

const GlobalStatsBar = () => {
  const classes = useStyles();

  const [stats, setStats] = useState<components['schemas']['ServiceStatus'] | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>('stats_global');

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
    const data: components['schemas']['ServiceStatus'] = await response.json();

    setStats(data);
  }, []);

  useEffect(() => {
    fetchStats();
    const fetchStatsInterval = setInterval(fetchStats, FETCH_STATS_INTERVAL);

    return () => {
      clearInterval(fetchStatsInterval);
    };
  }, [fetchStats]);

  const handleTabChange = (event: React.ChangeEvent<{}>, newValue: TabType) => {
    setActiveTab(newValue);
  };

  if (!stats) {
    return null;
  }

  const emptyStats = {
    peers_total: 0,
    peers_active: 0,
    traffic_up: '0 B',
    traffic_down: '0 B',
    speed_down: '0 Bps',
    speed_up: '0 Bps'
  };

  const activeStats = stats?.[activeTab];

  const displayStats = activeStats
    ? {
      ...activeStats,
      traffic_up: convertBytes(activeStats.traffic_up ?? 0),
      traffic_down: convertBytes(activeStats.traffic_down ?? 0),
      speed_up: `${convertBytes(activeStats.speed_up ?? 0)}ps`,
      speed_down: `${convertBytes(activeStats.speed_down ?? 0)}ps`
    }
    : emptyStats;
  const { peers_active, peers_total, traffic_up, traffic_down, speed_up, speed_down } = displayStats;

  return (
    <div className={classes.root}>
      <div className={classes.tabsContainer}>
        <Tabs
          value={activeTab}
          onChange={handleTabChange}
          indicatorColor="primary"
          textColor="primary"
          variant="fullWidth"
          className={classes.tabs}
        >
          {TABS.map((tab) => (
            <Tab key={tab.value} value={tab.value} label={tab.label} />
          ))}
        </Tabs>
      </div>

      <div className={classes.statsContent}>
        <div className={classes.row}>
          <span>Total peers:</span>
          <span>{peers_total}</span>
        </div>

        <div className={classes.row}>
          <span>Peers active:</span>
          <span>{peers_active}</span>
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
          <span>{speed_up}</span>
        </div>

        <div className={classes.row}>
          <span>Downstream speed:</span>
          <span>{speed_down}</span>
        </div>
      </div>
    </div>
  );
};

export default GlobalStatsBar;
