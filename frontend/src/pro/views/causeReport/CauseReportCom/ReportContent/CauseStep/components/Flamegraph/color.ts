import Color from 'color'
import color from 'tinycolor2'
export interface FlamegraphPalette {
  name: string
  goodColor: Color
  neutralColor: Color
  badColor: Color

  colors: Color[]
}
export const DefaultPalette: FlamegraphPalette = {
  name: 'Default',
  // green
  goodColor: Color.rgb(0, 170, 0),
  // grey
  neutralColor: Color.rgb(148, 142, 142),
  // red
  badColor: Color.rgb(200, 0, 0),

  colors: [
    Color.hsl(24, 69, 60),
    Color.hsl(34, 65, 65),
    Color.hsl(194, 52, 61),
    Color.hsl(163, 45, 55),
    Color.hsl(211, 48, 60),
    Color.hsl(246, 40, 65),
    Color.hsl(305, 63, 79),
    Color.hsl(47, 100, 73),

    Color.rgb(183, 219, 171),
    Color.rgb(244, 213, 152),
    Color.rgb(78, 146, 249),
    Color.rgb(249, 186, 143),
    Color.rgb(242, 145, 145),
    Color.rgb(130, 181, 216),
    Color.rgb(229, 168, 226),
    Color.rgb(174, 162, 224),
    Color.rgb(154, 196, 138),
    Color.rgb(242, 201, 109),
    Color.rgb(101, 197, 219),
    Color.rgb(249, 147, 78),
    Color.rgb(234, 100, 96),
    Color.rgb(81, 149, 206),
    Color.rgb(214, 131, 206),
    Color.rgb(128, 110, 183),
  ],
}
export const ColorBlindPalette: FlamegraphPalette = {
  ...DefaultPalette,

  name: 'Color Blind',
  goodColor: Color.rgb(26, 133, 255),
  neutralColor: Color.rgb(148, 142, 142),
  badColor: Color.rgb(220, 50, 32),
}

export default function murmurhashThree32GC(key: string, seed = 0) {
  let remainder
  let bytes
  let h1
  let h1b
  let c1
  let c2
  let k1
  let i

  remainder = key.length & 3 // key.length % 4
  bytes = key.length - remainder
  h1 = seed
  c1 = 0xcc9e2d51
  c2 = 0x1b873593
  i = 0

  while (i < bytes) {
    k1 =
      (key.charCodeAt(i) & 0xff) |
      ((key.charCodeAt(++i) & 0xff) << 8) |
      ((key.charCodeAt(++i) & 0xff) << 16) |
      ((key.charCodeAt(++i) & 0xff) << 24)
    ++i

    k1 = ((k1 & 0xffff) * c1 + ((((k1 >>> 16) * c1) & 0xffff) << 16)) & 0xffffffff
    k1 = (k1 << 15) | (k1 >>> 17)
    k1 = ((k1 & 0xffff) * c2 + ((((k1 >>> 16) * c2) & 0xffff) << 16)) & 0xffffffff

    h1 ^= k1
    h1 = (h1 << 13) | (h1 >>> 19)
    h1b = ((h1 & 0xffff) * 5 + ((((h1 >>> 16) * 5) & 0xffff) << 16)) & 0xffffffff
    h1 = (h1b & 0xffff) + 0x6b64 + ((((h1b >>> 16) + 0xe654) & 0xffff) << 16)
  }

  k1 = 0

  switch (remainder) {
    case 3:
      k1 ^= (key.charCodeAt(i + 2) & 0xff) << 16
    // fall through
    case 2:
      k1 ^= (key.charCodeAt(i + 1) & 0xff) << 8
    // fall through
    case 1:
      k1 ^= key.charCodeAt(i) & 0xff
    // fall through
    default:
      k1 = ((k1 & 0xffff) * c1 + ((((k1 >>> 16) * c1) & 0xffff) << 16)) & 0xffffffff
      k1 = (k1 << 15) | (k1 >>> 17)
      k1 = ((k1 & 0xffff) * c2 + ((((k1 >>> 16) * c2) & 0xffff) << 16)) & 0xffffffff
      h1 ^= k1
  }

  h1 ^= key.length

  h1 ^= h1 >>> 16
  h1 = ((h1 & 0xffff) * 0x85ebca6b + ((((h1 >>> 16) * 0x85ebca6b) & 0xffff) << 16)) & 0xffffffff
  h1 ^= h1 >>> 13
  h1 = ((h1 & 0xffff) * 0xc2b2ae35 + ((((h1 >>> 16) * 0xc2b2ae35) & 0xffff) << 16)) & 0xffffffff
  h1 ^= h1 >>> 16

  return h1 >>> 0
}
function spyToRegex(spyName: any): RegExp {
  // eslint-disable-next-line default-case
  switch (spyName) {
    case 'dotnetspy':
      return /^(?<packageName>.+)\.(.+)\.(.+)\(.*\)$/
    // TODO: come up with a clever heuristic
    case 'ebpfspy':
      return /^(?<packageName>.+)$/
    // tested with pyroscope stacktraces here: https://regex101.com/r/99KReq/1
    case 'gospy':
      return /^(?<packageName>.*?\/.*?\.|.*?\.|.+)(?<functionName>.*)$/
    // assume scrape is golang, since that's the only language we support right now
    case 'scrape':
      return /^(?<packageName>.*?\/.*?\.|.*?\.|.+)(?<functionName>.*)$/
    case 'phpspy':
      return /^(?<packageName>(.*\/)*)(?<filename>.*\.php+)(?<line_info>.*)$/
    case 'pyspy':
      return /^(?<packageName>(.*\/)*)(?<filename>.*\.py+)(?<line_info>.*)$/
    case 'rbspy':
      return /^(?<packageName>(.*\/)*)(?<filename>.*\.rb+)(?<line_info>.*)$/
    case 'nodespy':
      return /^(\.\/node_modules\/)?(?<packageName>[^/]*)(?<filename>.*\.?(jsx?|tsx?)?):(?<functionName>.*):(?<line_info>.*)$/
    case 'tracing':
      return /^(?<packageName>.+?):.*$/
    case 'javaspy':
      // TODO: we might want to add ? after groups
      return /^(?<packageName>.+\/)(?<filename>.+\.)(?<functionName>.+)$/
    case 'pyroscope-rs':
      return /^(?<packageName>[^::]+)/
    case 'unknown':
      return /^(?<packageName>.+)$/
  }

  return /^(?<packageName>.+)$/
}
export function getPackageNameFromStackTrace(spyName: any, stackTrace: string) {
  if (stackTrace.length === 0) {
    return stackTrace
  }
  const regexp = spyToRegex('')
  const fullStackGroups = stackTrace.match(regexp)
  if (fullStackGroups && fullStackGroups.groups) {
    return fullStackGroups.groups['packageName'] || ''
  }
  return stackTrace
}

export function colorBasedOnPackageName(name: string) {
  const hash = murmurhashThree32GC(name, 0)
  const colorIndex = hash % DefaultPalette.colors.length
  const baseClr = DefaultPalette.colors[colorIndex]
  if (!baseClr) {
    console.warn('Could not calculate color. Defaulting to the first one')
    // We assert to Color since the first position is always available
    return DefaultPalette.colors[0]
  }

  return baseClr
}

export function getBarColorByValue(value: number, totalTicks: number, rangeMin = 0, rangeMax = 1) {
  //  / (rangeMax - rangeMin) here so when you click a bar it will adjust the top (clicked)bar to the most 'intense' color
  const intensity = Math.min(1, value / totalTicks / (rangeMax - rangeMin))
  const h = 50 - 50 * intensity
  const l = 65 + 7 * intensity

  return color({ h, s: 100, l })
}
