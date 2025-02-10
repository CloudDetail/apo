import Color from 'color'
import murmurhashThree32GC from './murmur3'
// signleFlame
export const signleFlame = {
  step: 4,
  nameIndex: 3,
  //距离上一个结束值距离多少
  getOffset: (level: number[], index: number) => level[index],
  //总样本数
  getTotalSamples: (level: number[], index: number) => level[index + 1],
  //自身样本数
  getSelfSamples: (level: number[], index: number) => level[index + 2],
  //function name
  getFuncNameIndex: (level: number[], j: number) => level[j + 3],
}
export const DefaultPalette = {
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
export function getPackageNameFromStackTrace(stackTrace: string) {
  if (stackTrace.length === 0) {
    return stackTrace
  }
  const regexp = /^(?<packageName>.+)$/
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
export const getColor = (funcName: string) => {
  return colorBasedOnPackageName(getPackageNameFromStackTrace(funcName) || '')
}
