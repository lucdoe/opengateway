export default function CheckboxSelector({
  name,
  fields,
  onChange,
  checked,
}: Readonly<{
  name: string
  fields: string[]
  onChange: (fieldName: string, isChecked: boolean) => void
  checked: {[key: string]: boolean}
}>) {
  return (
    <div className='mt-6'>
      <fieldset>
        <legend className='text-sm font-medium text-gray-900'>{name}</legend>
        <div className='mt-2 divide-y divide-gray-200 border-b border-t border-gray-200 max-w-32'>
          {fields.map((field) => (
            <div key={field} className='relative flex items-start py-1'>
              <div className='min-w-0 flex-1 text-sm leading-6'>
                <label htmlFor={field} className='select-none text-gray-700'>
                  {field}
                </label>
              </div>
              <div className='ml-3 flex h-6 items-center'>
                <input
                  id={field}
                  name={field}
                  type='checkbox'
                  checked={checked[field]}
                  onChange={(e) => onChange(field, e.target.checked)}
                  className='h-4 w-4 rounded border-gray-300 text-green-400 focus:ring-green-500'
                />
              </div>
            </div>
          ))}
        </div>
      </fieldset>
    </div>
  )
}
