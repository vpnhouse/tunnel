import { $globalLoading, setGlobalLoading } from './index';

$globalLoading.on(setGlobalLoading, (_, res) => res);
