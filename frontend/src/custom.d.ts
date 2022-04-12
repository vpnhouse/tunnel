declare module '*.svg' {
  const content: string;
  export default content;
}

declare module '*.png'

interface Window {
  wireguard: {
    generateKeypair: () => ({publicKey: string, privateKey: string})
  };
}

declare type Entries<T> = {
  [K in keyof Required<T>]: [Required<K>, Required<T>[K]];
}[keyof Required<T>][];
