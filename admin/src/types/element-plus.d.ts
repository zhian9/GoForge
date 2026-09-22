// element-plus 的中文语言包是 .mjs 文件，TypeScript 找不到它的类型声明，
// 直接 import 会报 TS7016（隐式 any）。这里补一个最小声明。
// 用户端 web/src/types/element-plus.d.ts 有同一份。
declare module 'element-plus/dist/locale/zh-cn.mjs' {
  const locale: Record<string, unknown>
  export default locale
}
