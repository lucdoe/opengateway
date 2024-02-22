export interface Field {
  name?: string
  fieldtype: 'text' | 'textarea' | 'toggle' | 'checkbox'
  label: string
  placeholder?: string
  fields?: string[]
  id?: string
}

export interface FormData {
  [key: string]: any
}

export interface Endpoint {
  id: string
  name: string
  methods: string[]
  path: string
  enabled: boolean
  tags: string[]
}

export interface Service {
  id: string
  name: string
  protocols: string[]
  host: string
  port: string
  enabled: boolean
  tags: string[]
}

export interface FieldInputProps {
  field: Field
  formData: FormData
  handleInputChange: (fieldName: string, value: any) => void
}

export interface DynamicAddComponentProps {
  fields: Field[]
  name: string
  endpoint: string
  method: 'GET' | 'PUT' | 'POST' | 'DELETE'
  initialValues?: FormData
}

export const endpointsFields = [
  {fieldtype: 'toggle', label: 'Enabled', value: 'true'},
  {fieldtype: 'text', label: 'Name', placeholder: 'Test Name'},

  {fieldtype: 'text', label: 'Path', placeholder: '/test'},
  {
    fieldtype: 'checkbox',
    label: 'Methods',
    name: 'Methods',
    fields: ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS'],
  },
  {
    fieldtype: 'text',
    label: 'Tags',
    placeholder: 'test, example, tag',
  },
]

export const servicesFields = [
  {fieldtype: 'toggle', label: 'Enabled', value: 'true'},
  {},
  {fieldtype: 'text', label: 'Name', placeholder: 'Test Name'},
  {
    fieldtype: 'checkbox',
    label: 'Protocols',
    name: 'Protocols',
    fields: ['http', 'https', 'grpc', 'grpcs'],
  },
  {fieldtype: 'text', label: 'Host', placeholder: 'localhost'},
  {
    fieldtype: 'text',
    label: 'Port',
    placeholder: '8080',
  },

  {fieldtype: 'text', label: 'Tags', placeholder: 'test, tag'},
]
