'use client'
import {useDynamicForm} from '@/hooks/useDynamicForm'
import {DynamicAddComponentProps} from '@/interfaces/input'
import {formatRequest} from '@/utils/formatRequest'
import FieldInputSwitch from './FieldInputSwitch'
import SubmitBtn from './SubmitBtn'

export default function DynamicAddComponent({
  fields,
  name,
  endpoint,
  method,
  initialValues,
}: Readonly<DynamicAddComponentProps>) {
  const {formData, handleInputChange} = useDynamicForm(fields, initialValues)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const response = await fetch(endpoint, {
        method,
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(formatRequest(formData)),
      })
      if (response.ok) {
        window.location.href = `/${endpoint.split('/')[4]}`
      } else {
        console.error('Failed to submit:', await response.text())
      }
    } catch (error) {
      console.error('Network error:', error)
    }
  }

  return (
    <form onSubmit={handleSubmit} className=''>
      {fields.map((field) => (
        <FieldInputSwitch
          key={`${field.fieldtype}-${field.label}`}
          field={field}
          formData={formData as FormData}
          handleInputChange={handleInputChange}
        />
      ))}

      <SubmitBtn name={name} />
    </form>
  )
}
