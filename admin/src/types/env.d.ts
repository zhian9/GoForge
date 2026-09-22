/// <reference types="vite/client" />

// SkuList 用 import.meta.env 拼上传地址，需要 Vite 的环境变量类型。
interface ImportMetaEnv {
  /** 后端网关根地址，未配置时使用同源相对路径（/api） */
  readonly VITE_API_BASE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
