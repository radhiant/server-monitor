/**
 * Tree-shaken ECharts entry point.
 *
 * Importing the full `echarts` package pulls ~1MB of chart types we never use.
 * Here we register only the pieces this dashboard actually renders, which keeps
 * the vendor chunk small enough to load instantly on a NOC display.
 */
import * as echarts from 'echarts/core'
import { LineChart, PieChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([
  LineChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  CanvasRenderer,
])

// Type-only re-export: erased at build time, costs nothing at runtime.
export type { EChartsOption } from 'echarts'
export type ECharts = echarts.ECharts

export { echarts }
