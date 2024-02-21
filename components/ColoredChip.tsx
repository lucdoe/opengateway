function DynamicColoredSpan({
  color,
  text,
  withDot,
}: Readonly<{color: string; text: string; withDot: boolean}>) {
  const colorVariants: {[key: string]: string} = {
    red: 'bg-red-100 text-red-700 ring-red-600',
    yellow: 'bg-yellow-100 text-yellow-700 ring-yellow-600',
    green: 'bg-green-100 text-green-700 ring-green-600',
    blue: 'bg-blue-100 text-blue-700 ring-blue-600',
  }

  return (
    <span
      className={`inline-flex gap-1 items-center rounded-md px-2 py-1 text-xs font-medium ${colorVariants[color]} ring-1 ring-inset`}>
      {withDot && (
        <svg
          className={`h-1.5 w-1.5 fill-${color}-500`}
          viewBox='0 0 6 6'
          aria-hidden='true'>
          <circle cx={3} cy={3} r={3} fill={color} />
        </svg>
      )}
      {text}
    </span>
  )
}

export default function ColoredChip({
  text,
  color,
  withDot,
}: Readonly<{text: string; color: string; withDot: boolean}>) {
  return <DynamicColoredSpan color={color} text={text} withDot={withDot} />
}
