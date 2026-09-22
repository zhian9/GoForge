/// <reference types="vite/client" />

// 网关地址等环境变量需要 Vite 的类型声明，否则 import.meta.env 取不到类型
interface ImportMetaEnv {
  /** 后端网关根地址，未配置时使用同源相对路径（/api） */
  readonly VITE_API_BASE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
