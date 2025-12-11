// This file is a compatibility layer for makeStyles migration
import { SxProps, Theme } from '@mui/material/styles';

export default function useStyles(_props?: Record<string, unknown>) {
  return new Proxy({} as Record<string, string>, {
    get: (_, prop) => String(prop)
  });
}

export type { SxProps, Theme };
