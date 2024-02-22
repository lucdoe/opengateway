const apiURL = 'http://localhost:3001/api'

export const apiRoutes = {
  endpoints: {
    all: `${apiURL}/endpoints`,
    one: (id: string) => `${apiURL}/endpoints/${id}`,
  },
  services: {
    all: `${apiURL}/services`,
    one: (id: string) => `${apiURL}/services/${id}`,
  },
}
