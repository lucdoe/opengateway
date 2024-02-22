import {ChangeEvent, useEffect, useState} from 'react'

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

export const useDynamicForm = (
  initialFields: Field[],
  initialValues?: FormData,
) => {
  const [formData, setFormData] = useState<FormData>(initialValues ?? {})

  useEffect(() => {
    if (initialValues) {
      setFormData(initialValues)
    }
  }, [initialValues])

  const handleInputChange = (
    fieldName: string,
    eventOrValue:
      | ChangeEvent<HTMLInputElement>
      | ChangeEvent<HTMLTextAreaElement>
      | boolean,
  ) => {
    let value: any
    const isEvent = eventOrValue instanceof Event

    if (isEvent) {
      const target = eventOrValue.target as HTMLInputElement
      const isCheckbox = target.type === 'checkbox'
      value = isCheckbox ? target.checked : target.value
    } else {
      value = eventOrValue
    }

    setFormData((prevFormData) => ({
      ...prevFormData,
      [fieldName.toLowerCase()]: value,
    }))
  }

  return {formData, handleInputChange}
}
