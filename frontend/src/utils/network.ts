import type { NetworkInterface } from '@/types/metrics'

/**
 * Interfaces that exist on every Linux host but never carry the traffic an
 * operator cares about: loopback, bridges, container veth pairs, VPN taps.
 * A Docker host reports dozens of these and they sort ahead of the real NIC.
 */
const VIRTUAL_PREFIXES = /^(lo|docker|br-|veth|virbr|vnet|tun|tap|cni|flannel|kube|dummy|bond-slave)/i

export function isVirtualInterface(name: string): boolean {
  return VIRTUAL_PREFIXES.test(name)
}

function lifetimeBytes(iface: NetworkInterface): number {
  return (iface.total_rx_bytes || 0) + (iface.total_tx_bytes || 0)
}

/**
 * The interface the dashboard should show by default.
 *
 * Picking `network[0]` landed on `lo` on every one of these hosts, so the
 * network card and chart read a permanent 0.00 MB/s — worthless on a wall
 * display. Prefer a physical NIC, ranked by lifetime traffic.
 */
export function pickPrimaryInterface(
  interfaces: NetworkInterface[] | undefined,
): NetworkInterface | null {
  if (!interfaces || interfaces.length === 0) return null

  const physical = interfaces.filter(i => !isVirtualInterface(i.name))
  const pool = physical.length > 0 ? physical : interfaces

  return pool.reduce((best, iface) => (lifetimeBytes(iface) > lifetimeBytes(best) ? iface : best))
}

/**
 * Selector ordering: real NICs first, each group by lifetime traffic, so the
 * useful entries sit at the top of a list that can run to forty interfaces.
 */
export function sortInterfacesForDisplay(
  interfaces: NetworkInterface[] | undefined,
): NetworkInterface[] {
  return [...(interfaces || [])].sort((a, b) => {
    const virtualA = isVirtualInterface(a.name) ? 1 : 0
    const virtualB = isVirtualInterface(b.name) ? 1 : 0
    if (virtualA !== virtualB) return virtualA - virtualB
    return lifetimeBytes(b) - lifetimeBytes(a)
  })
}

/** Resolve the interface to render: the chosen one, else the best default. */
export function resolveInterface(
  interfaces: NetworkInterface[] | undefined,
  selectedName: string,
): NetworkInterface | null {
  const found = interfaces?.find(i => i.name === selectedName)
  return found || pickPrimaryInterface(interfaces)
}
