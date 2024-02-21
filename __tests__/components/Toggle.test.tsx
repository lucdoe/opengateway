import {render, screen} from '@testing-library/react'
import {test} from 'vitest'
import Toggle from '../../components/Toggle'

test('Toggle: renders correctly', ({expect}) => {
  const mockOnChange = () => {}
  render(<Toggle label='Test Label' onChange={mockOnChange} clicked={false} />)

  expect(screen.getByText('Test Label: Disabled')).toBeDefined()
})

test('Toggle: displays correct label when clicked is true', ({expect}) => {
  const mockOnChange = () => {}
  render(<Toggle label='Test Label' onChange={mockOnChange} clicked={true} />)

  expect(screen.getByText('Test Label: Active')).toBeDefined()
})
