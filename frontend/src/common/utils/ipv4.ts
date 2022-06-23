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

export const PRIVATE_SUBNETS = ['10.0.0.0/8', '172.16.0.0/12', '192.168.0.0/16'];

export function isSubnetPrivate(subnetInteger: number): boolean {
  return PRIVATE_SUBNETS.some((value) => {
    const [privateSubnetString, bitsString] = value.split('/');

    const privateSubnetInteger = ip4ToInt(privateSubnetString);
    const mask = maskFromBits(+bitsString);

    return (subnetInteger & mask) === privateSubnetInteger;
  });
}
