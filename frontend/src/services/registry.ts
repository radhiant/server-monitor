import type { ServerRegistryItem } from '@/types/server'

export async function fetchServerRegistry(): Promise<ServerRegistryItem[]> {
  try {
    const res = await fetch('/servers.json')
    if (!res.ok) {
      throw new Error(`Failed to load servers.json: ${res.statusText}`)
    }
    const data: ServerRegistryItem[] = await res.json()
    return data
  } catch (err) {
    console.warn('Unable to load /servers.json, falling back to local default:', err)
    return [
      {
        id: 'local',
        name: 'Local Node',
        description: 'Default Local Monitoring Agent',
        baseUrl: '/servers/local'
      }
    ]
  }
}
