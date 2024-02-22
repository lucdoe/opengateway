export default function FieldLabel({label}: Readonly<{label: string}>) {
  return (
    <label
      htmlFor={label}
      className='pl-1 block text-sm font-medium leading-6 text-gray-900'>
      {label}
    </label>
  )
}
