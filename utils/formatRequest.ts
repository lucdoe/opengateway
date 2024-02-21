export function formatRequest(formData: any) {
  const protocols =
    formData.http || formData.https || formData.grpc || formData.grpcs
  const methods = 'POST' || 'PUT' || 'PATCH' || 'DELETE' || 'GET' || 'OPTIONS'

  if (protocols) {
    formData.protocols = {
      http: formData.http || false,
      https: formData.https || false,
      grpc: formData.grpc || false,
      grpcs: formData.grpcs || false,
    }
    delete formData.http
    delete formData.https
    delete formData.grpc
    delete formData.grpcs
  }

  if (methods) {
    formData.methods = {
      POST: formData.POST || false,
      PUT: formData.PUT || false,
      PATCH: formData.PATCH || false,
      DELETE: formData.DELETE || false,
      GET: formData.GET || false,
      OPTIONS: formData.OPTIONS || false,
    }
    delete formData.POST
    delete formData.PUT
    delete formData.PATCH
    delete formData.DELETE
    delete formData.GET
    delete formData.OPTIONS
  }

  if (formData.Tags) {
    formData.Tags = formData.Tags.split(',').map((tag: string) => tag.trim())
  }

  const result = Object.keys(formData).reduce(
    (acc, key) => ({
      ...acc,
      [key.toLowerCase()]: formData[key],
    }),
    {},
  )

  return result
}
