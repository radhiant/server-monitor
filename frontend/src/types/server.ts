export interface ServerRegistryItem {
  id: string
  name: string
  description?: string
  baseUrl: string
  apiToken?: string
  /** Physical role, e.g. "Web Server", "Database". Shown on the fleet card. */
  role?: string
  /** Physical placement, e.g. "Rack A-03 / U12". Overlaid on the photo. */
  location?: string
  /** Default photo of the physical machine, served from /public. */
  image?: string
}
