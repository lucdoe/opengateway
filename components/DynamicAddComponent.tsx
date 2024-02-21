'use client'
import {useDynamicForm} from '@/hooks/useDynamicForm'
import {formatRequest} from '@/utils/formatRequest'
import CheckboxSelector from './CheckBoxSelector'
import Toggle from './Toggle'

export default function DynamicAddComponent({
  fields,
  name,
  endpoint,
  method,
}: Readonly<{fields: any[]; name: string; endpoint: string; method: string}>) {
  const {formData, handleInputChange} = useDynamicForm(fields)

  const handleSubmit = (e: {preventDefault: () => void}) => {
    e.preventDefault()

    const requestData = formatRequest(formData)

    const request = fetch(endpoint, {
      method: method,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestData),
    })

    // check if it was successful and then redirect to the all page of endpoint
    request.then((res) => {
      if (res.status === 200) {
        window.location.href = `/${endpoint.split('/')[4]}`
      }
    })
  }

  return (
    <form onSubmit={handleSubmit} className=''>
      {fields.map((field, index) => {
        const fieldKey = `${field.fieldtype}-${field.label}-${index}`
        if (field.fieldtype === 'text') {
          return (
            <div
              key={fieldKey}
              className={
                index === 0
                  ? 'mt-8 min-w-64 max-w-96'
                  : 'mt-4 min-w-64 max-w-96'
              }>
              <label
                htmlFor={field.label}
                className='pl-1 block text-sm font-medium leading-6 text-gray-900'>
                {field.label}
              </label>
              <div className='mt-1'>
                <input
                  type='text'
                  name={field.label}
                  id={field.label}
                  className='block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6'
                  placeholder={field.placeholder}
                  value={formData[field.label] || ''}
                  onChange={(e) => handleInputChange(field.label, e)}
                />
              </div>
            </div>
          )
        } else if (field.fieldtype === 'textarea') {
          return (
            <div key={fieldKey} className='mt-4'>
              <label
                htmlFor={field.label}
                className='pl-1 block text-sm font-medium leading-6 text-gray-900'>
                {field.label}
              </label>
              <textarea
                name={field.label}
                id={field.label}
                placeholder={field.placeholder}
                className='block w-full mt-1 rounded-md border-gray-300 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50'
                value={formData[field.label] || ''}
                onChange={(e) => handleInputChange(field.label, e)}
              />
            </div>
          )
        } else if (field.fieldtype === 'toggle') {
          return (
            <div key={fieldKey} className='pl-1 mt-4'>
              <Toggle
                label={field.label}
                clicked={formData[field.label] || false}
                onChange={(isChecked) =>
                  handleInputChange(field.label, isChecked)
                }
              />
            </div>
          )
        } else if (field.fieldtype === 'checkbox') {
          return (
            <div key={fieldKey} className='pl-1 mt-4'>
              <CheckboxSelector
                fields={field.fields}
                name={field.name}
                onChange={(fieldName, isChecked) =>
                  handleInputChange(fieldName, isChecked)
                }
                checked={formData}
              />
            </div>
          )
        } else {
          return <div key='empty'></div>
        }
      })}

      <button
        type='submit'
        className='mt-8 max-w-48 rounded-md bg-indigo-600 px-3.5 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'>
        {name}
      </button>
    </form>
  )
}
