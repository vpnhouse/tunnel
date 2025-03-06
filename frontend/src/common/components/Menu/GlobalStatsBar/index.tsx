import React, { useCallback, useEffect, useState } from 'react';
import { Tabs, Tab } from '@material-ui/core';

import { fetchData } from '@root/store/utils';
import { GLOBAL_STATS } from '@constants/apiPaths';
import { GlobalStats, GlobalStatsResponse, TabType } from '@common/components/Menu/GlobalStatsBar/types';
import { BYTES_MEASURE_LIMITS, FETCH_STATS_INTERVAL, TABS } from '@common/components/Menu/GlobalStatsBar/constant';

import useStyles from './styles';

const GlobalStatsBar = () => {
  const classes = useStyles();

  const [stats, setStats] = useState<GlobalStats | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>('all');

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
      traffic_up_speed: `${convertBytes(data.traffic_up_speed)}ps`,
      traffic_down_speed: `${convertBytes(data.traffic_down_speed)}ps`
    });
  }, [convertBytes]);

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

  const getDisplayStats = () => {
    if (activeTab === 'iprose') {
      return {
        peers_total: 0,
        peers_connected: 0,
        peers_active_1h: 0,
        peers_active_1d: 0,
        traffic_up: '0 B',
        traffic_down: '0 B',
        traffic_up_speed: '0 Bps',
        traffic_down_speed: '0 Bps'
      };
    }

    return stats;
  };

  const displayStats = getDisplayStats();
  const { peers_active_1h, peers_active_1d, peers_connected, peers_total, traffic_up, traffic_down, traffic_up_speed, traffic_down_speed } = displayStats;

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
    </div>
  );
};

export default GlobalStatsBar;
