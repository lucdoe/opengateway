import {FormData} from '@/interfaces/input'
import {ChangeEvent, useEffect, useState} from 'react'

export const useDynamicForm = (initialValues?: FormData) => {
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
      [fieldName]: value,
    }))
  }

  return {formData, handleInputChange}
}
