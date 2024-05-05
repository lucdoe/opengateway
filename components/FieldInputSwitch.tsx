import {FieldInputProps} from '@/interfaces/input'
import CheckboxSelector from './CheckBoxSelector'
import FieldLabel from './FieldLabel'
import Toggle from './Toggle'

export default function FieldInputSwitch({
  field,
  formData,
  handleInputChange,
}: Readonly<FieldInputProps>) {
  const commonFieldProps = {
    id: field.id ?? field.label,
    className:
      'block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6',
    placeholder: field.placeholder,
  }

  const determineInputBasedOnFieldtype = () => {
    switch (field.fieldtype) {
      case 'text':
        return (
          <div className='mt-4 min-w-64 max-w-96'>
            <FieldLabel label={field.label} />
            <input
              type='text'
              {...commonFieldProps}
              value={formData[field.label.toLowerCase()] || ''}
              onChange={(e) =>
                handleInputChange(field.label.toLowerCase(), e.target.value)
              }
            />
          </div>
        )

      case 'textarea':
        return (
          <div className='mt-4 min-w-64 max-w-96'>
            <FieldLabel label={field.label} />
            <textarea
              {...commonFieldProps}
              value={formData[field.label.toLowerCase()] || ''}
              onChange={(e) =>
                handleInputChange(field.label.toLowerCase(), e.target.value)
              }
            />
          </div>
        )

      case 'toggle':
        return (
          <div className='pl-1 mt-4'>
            <Toggle
              label={field.label}
              clicked={Boolean(formData[field.label.toLowerCase()])}
              onChange={(isChecked) =>
                handleInputChange(field.label.toLowerCase(), isChecked)
              }
            />
          </div>
        )

      case 'checkbox':
        return (
          <div className='pl-1 mt-4'>
            <CheckboxSelector
              fields={field.fields ?? []}
              name={field.name ?? ''}
              onChange={(fieldName, isChecked) =>
                handleInputChange(fieldName, isChecked)
              }
              checked={formData[field.label.toLowerCase()] || {}}
            />
          </div>
        )

      default:
        return null
    }
  }

  return determineInputBasedOnFieldtype()
}
