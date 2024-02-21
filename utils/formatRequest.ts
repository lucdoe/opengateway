export function formatRequest(formData: any) {
  const protocols = ['http', 'https', 'grpc', 'grpcs']
  const methods = ['POST', 'PUT', 'PATCH', 'DELETE', 'GET', 'OPTIONS']

  if (!formData.enabled) {
    formData.enabled = false
  }

  formData.protocols = protocols.reduce((acc: string[], key: string) => {
    if (formData[key.toLowerCase()]) {
      acc.push(key.toLowerCase())
    }
    return acc
  }, [])

  formData.methods = methods.reduce((acc: string[], key: string) => {
    if (formData[key.toUpperCase()]) {
      acc.push(key.toUpperCase())
    }
    return acc
  }, [])

  methods.forEach((key) => {
    delete formData[key.toUpperCase()]
  })

  if (formData.Tags) {
    formData.Tags = formData.Tags.split(',').map((tag: string) => tag.trim())
  }

  const results: any = {}
  for (const key in formData) {
    results[key.toLowerCase()] = formData[key]
  }

  return results
}
