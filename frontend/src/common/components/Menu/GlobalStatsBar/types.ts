export interface GlobalStats {
  peers_total: number;
  peers_connected: number;
  peers_active_1h: number;
  peers_active_1d: number;
  traffic_rx: string;
  traffic_tx: string;
}

export interface GlobalStatsResponse {
  peers_total: number;
  peers_connected: number;
  peers_active_1h: number;
  peers_active_1d: number;
  traffic_rx: number;
  traffic_tx: number;
}
