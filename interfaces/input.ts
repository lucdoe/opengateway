interface Field {
  name?: string
  fieldtype: 'text' | 'textarea' | 'toggle' | 'checkbox'
  label: string
  placeholder?: string
  fields?: string[]
  id?: string
}

interface FormData {
  [key: string]: any
}

interface FieldInputProps {
  field: Field
  formData: FormData
  handleInputChange: (fieldName: string, value: any) => void
}

interface DynamicAddComponentProps {
  fields: Field[]
  name: string
  endpoint: string
  method: 'GET' | 'PUT' | 'POST' | 'DELETE'
  initialValues?: FormData
}
