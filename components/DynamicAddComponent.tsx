import CheckboxSelector from './CheckBoxSelector'
import Toggle from './Toggle'

export default function DynamicAddComponent({
  fields,
  name,
}: Readonly<{fields: any[]; name: string}>) {
  return (
    <div className=''>
      {fields.map((field, index) => {
        if (field.fieldtype === 'text') {
          return (
            <div
              key={field.label}
              className={
                index == 0 ? 'mt-8 min-w-64 max-w-96' : 'mt-4 min-w-64 max-w-96'
              }>
              <label
                htmlFor='text'
                className='pl-1 block text-sm font-medium leading-6 text-gray-900'>
                {field.label}
              </label>
              <div className='mt-1'>
                <input
                  type='text'
                  name='text'
                  id='text'
                  className='block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6'
                  placeholder={field.placeholder}
                />
              </div>
            </div>
          )
        } else if (field.fieldtype === 'textarea') {
          return (
            <textarea
              key={name}
              placeholder={field.label}
              value={field.value}
            />
          )
        } else if (field.fieldtype === 'select') {
          return (
            <select key={name} value={field.value}>
              <option value=''>{field.label}</option>
              <option value='1'>Option 1</option>
              <option value='2'>Option 2</option>
            </select>
          )
        } else if (field.fieldtype === 'radio') {
          return (
            <label key={name}>
              <input type='radio' value={field.value} />
              {field.label}
            </label>
          )
        } else if (field.fieldtype === 'toggle') {
          return (
            <div key={field} className='pl-1 mt-4'>
              <Toggle key={field} />
            </div>
          )
        } else if (field.fieldtype === 'checkbox') {
          return (
            <div key={field} className='pl-1 mt-4'>
              <CheckboxSelector
                key={field}
                fields={field.fields}
                name={field.name}
              />
            </div>
          )
        } else {
          return <div key='empty'></div>
        }
      })}

      <button
        type='button'
        className='mt-8 max-w-48 rounded-md bg-indigo-600 px-3.5 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'>
        {name}
      </button>
    </div>
  )
}
