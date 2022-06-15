export function maskFromBits(bitsNumber: number): number {
  const bits: string[] = [];

  for (let i = 1; i <= 32; i++) {
    bits.push(i <= bitsNumber ? '1' : '0');
  }

  return parseInt(bits.join(''), 2);
}

export function ip4ToInt(ip: string) {
  return ip.split('.').reduce((int, oct) => (int << 8) + parseInt(oct, 10), 0) >>> 0;
}
