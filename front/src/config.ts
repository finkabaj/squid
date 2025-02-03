interface IConfig {
  API_URL: string
}

const config: IConfig = {
  API_URL: import.meta.env.VITE_API_URL,
}

export default config
