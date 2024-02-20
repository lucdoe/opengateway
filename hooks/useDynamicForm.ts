import {ChangeEvent, useState} from 'react'

interface FormData {
  [key: string]: any
}

interface Field {
  fieldtype: 'text' | 'textarea' | 'toggle' | 'checkbox'
  label: string
  placeholder?: string
  fields?: string[]
  name?: string
}

export const useDynamicForm = (initialFields: Field[]) => {
  const [formData, setFormData] = useState<FormData>({})

  const handleInputChange = (
    fieldName: string,
    eventOrValue:
      | ChangeEvent<HTMLInputElement>
      | ChangeEvent<HTMLTextAreaElement>
      | boolean,
  ) => {
    let value: any

    if (typeof eventOrValue === 'boolean') {
      value = eventOrValue
    } else if (
      (eventOrValue as ChangeEvent<HTMLInputElement>).target.type === 'checkbox'
    ) {
      value = (eventOrValue as ChangeEvent<HTMLInputElement>).target.checked
    } else {
      value = (eventOrValue as ChangeEvent<HTMLInputElement>).target.value
    }

    setFormData((prevFormData) => ({
      ...prevFormData,
      [fieldName]: value,
    }))
  }

  return {formData, handleInputChange}
}
